package spflib

import (
	"strings"
	"testing"

	"github.com/StackExchange/dnscontrol/pkg/dnsresolver"
)

func TestFlatten(t *testing.T) {
	res, err := dnsresolver.NewResolverPreloaded("testdata-dns1.json")
	if err != nil {
		t.Fatal(err)
	}
	rec, err := Parse(strings.Join([]string{"v=spf1",
		"ip4:198.252.206.0/24",
		"ip4:192.111.0.0/24",
		"include:_spf.google.com",
		"include:mailgun.org",
		"include:spf-basic.fogcreek.com",
		"include:mail.zendesk.com",
		"include:servers.mcsv.net",
		"include:sendgrid.net",
		"include:spf.mtasv.net",
		"~all"}, " "), res)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(rec.Print())
	rec = rec.Flatten("mailgun.org")
	//fmt.Println(rec.TXT())
	//fmt.Println(rec.TXTSplit("_spf%d.stackoverflow.com"))
	t.Log(rec.Print())
}

// each test is array of strings.
// first item is unsplit input
// next is @ spf record
// after that is alternating record fqdn and value
var splitTests = [][]string{
	{
		"simple",
		"v=spf1 -all",
		"v=spf1 -all",
	},
	{
		"longsimple",
		"v=spf1 include:a01234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789.com -all",
		"v=spf1 include:a01234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789.com -all",
	},
	{
		"long simple multipart",
		"v=spf1 include:a.com include:b.com include:12345678901234567890123456789000000000000000123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789.com -all",
		"v=spf1 include:a.com include:b.com include:12345678901234567890123456789000000000000000123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789.com -all",
	},
	{
		"overflow",
		"v=spf1 include:a.com include:b.com include:X12345678901234567890123456789000000000000000123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789.com -all",
		"v=spf1 include:a.com include:b.com include:_spf1.stackex.com -all",
		"_spf1.stackex.com",
		"v=spf1 include:X12345678901234567890123456789000000000000000123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789.com -all",
	},
	{
		"overflow all sign carries",
		"v=spf1 include:a.com include:b.com include:X12345678901234567890123456789000000000000000123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789.com ~all",
		"v=spf1 include:a.com include:b.com include:_spf1.stackex.com ~all",
		"_spf1.stackex.com",
		"v=spf1 include:X12345678901234567890123456789000000000000000123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789.com ~all",
	},
	{
		"really big",
		"v=spf1 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178" +
			" ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178" +
			" ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 -all",
		"v=spf1 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 include:_spf1.stackex.com -all",
		"_spf1.stackex.com",
		"v=spf1 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 include:_spf2.stackex.com -all",
		"_spf2.stackex.com",
		"v=spf1 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 include:_spf3.stackex.com -all",
		"_spf3.stackex.com",
		"v=spf1 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 ip4:200.192.169.178 -all",
	},
	{
		"too long to split",
		"v=spf1 include:a0123456789012345678901234567890123456789012345sssss6789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789.com -all",
		"v=spf1 include:a0123456789012345678901234567890123456789012345sssss6789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789.com -all",
	},
}

func TestSplit(t *testing.T) {

	for _, tst := range splitTests {
		t.Run(tst[0], func(t *testing.T) {
			rec, err := Parse(tst[1], nil)
			if err != nil {
				t.Fatal(err)
			}
			res := rec.TXTSplit("_spf%d.stackex.com")
			if res["@"] != tst[2] {
				t.Fatalf("Root record wrong. \nExp %s\ngot %s", tst[2], res["@"])
			}
			for i := 3; i < len(tst); i += 2 {
				fqdn := tst[i]
				exp := tst[i+1]
				if res[fqdn] != exp {
					t.Fatalf("Record %s.\nExp %s\ngot %s", fqdn, exp, res[fqdn])
				}
			}
		})
	}
}