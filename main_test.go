package main

import "testing"

func TestGetMessageTimestamp(t *testing.T) {
	m, ts := getMessageTimestamp("750: date=2019-11-28 time=09:33:52 logid=\"0000000013\" type=\"traffic\" subtype=\"forward\" level=\"notice\" vd=\"root\" eventtime=1574951632 srcip=192.168.6.101 srcname=\"S4Ns-Mac-mini-2-local\" srcport=61202 srcintf=\"internal\" srcintfrole=\"lan\" dstip=17.248.137.8 dstport=443 dstintf=\"wan2\" dstintfrole=\"wan\" poluuid=\"f4260370-15e9-51e9-19df-8084f3f618f7\" sessionid=47588464 proto=6 action=\"close\" policyid=154 policytype=\"policy\" service=\"HTTPS\" dstcountry=\"United States\" srccountry=\"Reserved\" trandisp=\"snat\" transip=192.168.1.21 transport=61202 duration=32 sentbyte=2054 rcvdbyte=8141 sentpkt=15 rcvdpkt=12 appcat=\"unscanned\" devtype=\"Mac\" devcategory=\"iOS Device\" osname=\"Mac OS X\" mastersrcmac=\"f0:18:98:e9:65:2a\" srcmac=\"f0:18:98:e9:65:2a\" srcserver=1")
	expectedMessage := "750: date=2019-11-28 time=09:33:52 logid=\"0000000013\" type=\"traffic\" subtype=\"forward\" level=\"notice\" vd=\"root\" eventtime=1574951632 srcip=192.168.6.101 srcname=\"S4Ns-Mac-mini-2-local\" srcport=61202 srcintf=\"internal\" srcintfrole=\"lan\" dstip=17.248.137.8 dstport=443 dstintf=\"wan2\" dstintfrole=\"wan\" poluuid=\"f4260370-15e9-51e9-19df-8084f3f618f7\" sessionid=47588464 proto=6 action=\"close\" policyid=154 policytype=\"policy\" service=\"HTTPS\" dstcountry=\"United States\" srccountry=\"Reserved\" trandisp=\"snat\" transip=192.168.1.21 transport=61202 duration=32 sentbyte=2054 rcvdbyte=8141 sentpkt=15 rcvdpkt=12 appcat=\"unscanned\" devtype=\"Mac\" devcategory=\"iOS Device\" osname=\"Mac OS X\" mastersrcmac=\"f0:18:98:e9:65:2a\" srcmac=\"f0:18:98:e9:65:2a\" srcserver=1"
	var expectedTimestamp int64
	expectedTimestamp = 1574951632000
	if m != expectedMessage {
		t.Fatalf("Message is Wrong!\nExpected: %s\nReceived: %s\n", expectedMessage, m)
	}
	if ts != expectedTimestamp {
		t.Fatalf("Timestamp is Wrong!\nExpected: %d\nReceived: %d\n", expectedTimestamp, ts)
	}
}
