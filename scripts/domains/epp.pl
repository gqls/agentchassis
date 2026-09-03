#!/usr/bin/perl
# Nominet EPP client (read-only commands).
#
# TLS is provided by `openssl s_client`; this script owns the RFC 5734 framing
# (4-byte big-endian length INCLUDING itself, then the XML). Perl in the target
# pod has no IO::Socket::SSL, hence the openssl transport.
#
# Credentials arrive on STDIN: line 1 = TAG, line 2 = password. They are never
# written to this file, a command line, or a process list.
#
# Usage: perl epp.pl <mode> [month]
#   login          login, print the result code, logout.
#                  THIS IS THE ALLOWLIST TEST: Nominet serves the greeting to any
#                  IP, so only a LOGIN proves the egress address is allowlisted.
#   list <YYYY-MM> login, list domains expiring in that month, logout.
use strict;
use warnings;
use IPC::Open2;

my $mode  = shift(@ARGV) // 'login';
my $month = shift(@ARGV) // '';

chomp(my $tag = <STDIN> // '');
chomp(my $pw  = <STDIN> // '');
die "no TAG/password on stdin\n" unless length($tag) && length($pw);

my ($R, $W);
my $pid = open2($R, $W,
    'openssl s_client -connect epp.nominet.org.uk:700 -4 -quiet 2>/dev/null');
binmode($R);
binmode($W);

sub rd {    # read exactly one EPP frame
    my $h = '';
    my $n = 0;
    while ($n < 4) {
        my $g = read($R, my $b, 4 - $n);
        return undef unless $g;
        $h .= $b;
        $n += $g;
    }
    my $len = unpack('N', $h) - 4;
    return '' if $len <= 0 || $len > 20_000_000;
    my $body = '';
    $n = 0;
    while ($n < $len) {
        my $g = read($R, my $b, $len - $n);
        return undef unless $g;
        $body .= $b;
        $n += $g;
    }
    return $body;
}

sub wr {
    my $x = shift;
    print $W pack('N', length($x) + 4), $x;
    $W->flush();
}

sub esc {
    my $s = shift;
    $s =~ s/&/&amp;/g;
    $s =~ s/</&lt;/g;
    $s =~ s/>/&gt;/g;
    return $s;
}

my $greeting = rd();
die "no greeting\n" unless defined $greeting;
print "GREETING_BYTES=" . length($greeting) . "\n";

wr(
qq{<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<epp xmlns="urn:ietf:params:xml:ns:epp-1.0"><command><login>
<clID>} . esc($tag) . qq{</clID><pw>} . esc($pw) . qq{</pw>
<options><version>1.0</version><lang>en</lang></options>
<svcs><objURI>urn:ietf:params:xml:ns:domain-1.0</objURI>
<objURI>urn:ietf:params:xml:ns:contact-1.0</objURI>
<objURI>urn:ietf:params:xml:ns:host-1.0</objURI>
<svcExtension><extURI>http://www.nominet.org.uk/epp/xml/std-list-1.0</extURI></svcExtension></svcs>
</login><clTRID>login-1</clTRID></command></epp>}
);

my $lr = rd() // '';
my ($code) = $lr =~ /<result code="(\d+)"/;
my ($msg)  = $lr =~ /<msg[^>]*>([^<]*)</;
print "LOGIN_CODE=" . ($code // '?') . "\n";
print "LOGIN_MSG="  . ($msg  // '?') . "\n";

if (($code // '') eq '1000' && $mode eq 'list' && $month) {
    wr(
qq{<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<epp xmlns="urn:ietf:params:xml:ns:epp-1.0"><command><info>
<list:list xmlns:list="http://www.nominet.org.uk/epp/xml/std-list-1.0">
<list:expiry>$month</list:expiry></list:list>
</info><clTRID>list-$month</clTRID></command></epp>}
    );
    my $resp = rd() // '';
    my ($lc) = $resp =~ /<result code="(\d+)"/;
    print "LIST_CODE=" . ($lc // '?') . "\n";
    # std-list-1.0's own element is <list:domainName>, NOT <domain:name>
    # (that belongs to the unrelated domain-1.0 schema) - the wrong tag
    # matched zero names on every real response with no error for two
    # weeks (2026-09-03). noDomains is the schema's own count; cross-check
    # against it so a future shape drift is loud, not a silent empty list.
    my @names = ($resp =~ /<list:domainName>([^<]+)<\/list:domainName>/g);
    print "DOMAIN\t$_\n" for @names;
    my ($claimed) = $resp =~ /noDomains="(\d+)"/;
    if (defined $claimed && $claimed != scalar(@names)) {
        print "PARSER_MISMATCH claimed=$claimed parsed=" . scalar(@names) . "\n";
    }
    unless (@names) {
        my ($m2) = $resp =~ /<msg[^>]*>([^<]*)</;
        print "LIST_MSG=" . ($m2 // '?') . "\n";
    }
}

wr(
qq{<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<epp xmlns="urn:ietf:params:xml:ns:epp-1.0"><command><logout/><clTRID>bye-1</clTRID></command></epp>}
);
rd();
close($W);
close($R);
waitpid($pid, 0);
