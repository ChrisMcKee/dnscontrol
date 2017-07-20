package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/pkg/nameservers"
	"github.com/StackExchange/dnscontrol/providers"
	_ "github.com/StackExchange/dnscontrol/providers/_all"
)

//go:generate go run build/generate/generate.go

var credsFile = flag.String("creds", "creds.json", "Provider credentials JSON file")
var jsonFile = flag.String("json", "", "File containing intermediate JSON")

var flagProviders = flag.String("providers", "", "Providers to enable (comma seperated list); default is all-but-bind. Specify 'all' for all (including bind)")
var domains = flag.String("domains", "", "Comma seperated list of domain names to include")

var interactive = flag.Bool("i", false, "Confirm or Exclude each correction before they run")

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()
	command := flag.Arg(0)
	if command == "version" {
		fmt.Println(versionString())
		return
	}

	var dnsConfig *models.DNSConfig

	switch command {
	case "create-domains":
		for _, domain := range dnsConfig.Domains {
			fmt.Println("*** ", domain.Name)
			for prov := range domain.DNSProviders {
				dsp, ok := dsps[prov]
				if !ok {
					log.Fatalf("DSP %s not declared.", prov)
				}
				if creator, ok := dsp.(providers.DomainCreator); ok {
					fmt.Println("  -", prov)
					err := creator.EnsureDomainExists(domain.Name)
					if err != nil {
						fmt.Printf("Error creating domain: %s\n", err)
					}
				}
			}
		}
	case "preview", "push":
	DomainLoop:
		for _, domain := range dnsConfig.Domains {
			if !shouldRunDomain(domain.Name) {
				continue
			}
			fmt.Printf("******************** Domain: %s\n", domain.Name)
			nsList, err := nameservers.DetermineNameservers(domain, 0, dsps)
			if err != nil {
				log.Fatal(err)
			}
			domain.Nameservers = nsList
			nameservers.AddNSRecords(domain)
			for prov := range domain.DNSProviders {
				dc, err := domain.Copy()
				if err != nil {
					log.Fatal(err)
				}
				shouldrun := shouldRunProvider(prov, dc, nonDefaultProviders)
				statusLbl := ""
				if !shouldrun {
					statusLbl = "(skipping)"
				}
				fmt.Printf("----- DNS Provider: %s... %s", prov, statusLbl)
				if !shouldrun {
					fmt.Println()
					continue
				}
				dsp, ok := dsps[prov]
				if !ok {
					log.Fatalf("DSP %s not declared.", prov)
				}
				corrections, err := dsp.GetDomainCorrections(dc)
				if err != nil {
					fmt.Println("ERROR")
					anyErrors = true
					fmt.Printf("Error getting corrections: %s\n", err)
					continue DomainLoop
				}
				totalCorrections += len(corrections)
				plural := "s"
				if len(corrections) == 1 {
					plural = ""
				}
				fmt.Printf("%d correction%s\n", len(corrections), plural)
				anyErrors = printOrRunCorrections(corrections, command) || anyErrors
			}
			if run := shouldRunProvider(domain.Registrar, domain, nonDefaultProviders); !run {
				continue
			}
			fmt.Printf("----- Registrar: %s\n", domain.Registrar)
			reg, ok := registrars[domain.Registrar]
			if !ok {
				log.Fatalf("Registrar %s not declared.", reg)
			}
			if len(domain.Nameservers) == 0 && domain.Metadata["no_ns"] != "true" {
				fmt.Printf("No nameservers declared; skipping registrar. Add {no_ns:'true'} to force.\n")
				continue
			}
			dc, err := domain.Copy()
			if err != nil {
				log.Fatal(err)
			}
			corrections, err := reg.GetRegistrarCorrections(dc)
			if err != nil {
				fmt.Printf("Error getting corrections: %s\n", err)
				anyErrors = true
				continue
			}
			totalCorrections += len(corrections)
			anyErrors = printOrRunCorrections(corrections, command) || anyErrors
		}
	default:
		log.Fatalf("Unknown command %s", command)
	}
	if os.Getenv("TEAMCITY_VERSION") != "" {
		fmt.Fprintf(os.Stderr, "##teamcity[buildStatus status='SUCCESS' text='%d corrections']", totalCorrections)
	}
	fmt.Printf("Done. %d corrections.\n", totalCorrections)
	if anyErrors {
		os.Exit(1)
	}
}

var reader = bufio.NewReader(os.Stdin)

func printOrRunCorrections(corrections []*models.Correction, command string) (anyErrors bool) {
	anyErrors = false
	if len(corrections) == 0 {
		return anyErrors
	}
	for i, correction := range corrections {
		fmt.Printf("#%d: %s\n", i+1, correction.Msg)
		if command == "push" {
			if *interactive {
				fmt.Print("Run? (Y/n): ")
				txt, err := reader.ReadString('\n')
				run := true
				if err != nil {
					run = false
				}
				txt = strings.ToLower(strings.TrimSpace(txt))
				if txt != "y" {
					run = false
				}
				if !run {
					fmt.Println("Skipping")
					continue
				}
			}
			err := correction.F()
			if err != nil {
				fmt.Println("FAILURE!", err)
				anyErrors = true
			} else {
				fmt.Println("SUCCESS!")
			}
		}
	}
	return anyErrors
}

func shouldRunProvider(p string, dc *models.DomainConfig, nonDefaultProviders []string) bool {
	if *flagProviders == "all" {
		return true
	}
	if *flagProviders == "" {
		for _, pr := range nonDefaultProviders {
			if pr == p {
				return false
			}
		}
		return true
	}
	for _, prov := range strings.Split(*flagProviders, ",") {
		if prov == p {
			return true
		}
	}
	return false
}

func shouldRunDomain(d string) bool {
	if *domains == "" {
		return true
	}
	for _, dom := range strings.Split(*domains, ",") {
		if dom == d {
			return true
		}
	}
	return false
}

// Version management. 2 Goals:
// 1. Someone who just does "go get" has at least some information.
// 2. If built with build.sh, more specific build information gets put in.
// Update the number here manually each release, so at least we have a range for go-get people.
var (
	SHA       = ""
	Version   = "0.1.0"
	BuildTime = ""
)

// printVersion prints the version banner.
func versionString() string {
	var version string
	if SHA != "" {
		version = fmt.Sprintf("%s (%s)", Version, SHA)
	} else {
		version = fmt.Sprintf("%s-dev", Version) //no SHA. '0.x.y-dev' indeicates it is run form source without build script.
	}
	if BuildTime != "" {
		i, err := strconv.ParseInt(BuildTime, 10, 64)
		if err == nil {
			tm := time.Unix(i, 0)
			version += fmt.Sprintf(" built %s", tm.Format(time.RFC822))
		}
	}
	return fmt.Sprintf("dnscontrol %s", version)
}