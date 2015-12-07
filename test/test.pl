#!/usr/bin/env perl

my $cache = [];
my ($a, $b, $c) = (4, 1, 0);

while ($a != 6 && $c < 32768) {
	$c++;
	($a, $b) = (4, 1);
	$cache = [];
	print 'Trying: ' . join(', ', ($a, $b, $c)) . "\n";
	recurse();
}

print join(', ', ($a, $b, $c)) . "\n";


sub recurse {
	my $key = ($a * 32768) + $b;
	if ($cache->[$key]) {
		$a = $cache->[$key];
		return;
	}

	if ($a) {
		if ($b) {
			my $tmp = $a;
			$b--;
			recurse();
			$b = $a;
			$a = $tmp;
			$a--;
			recurse();
		} else {
			$a--;
			$b = $c;
			recurse();
		}
	} else {
		$a = ($b + 1) % 32768;
	}
	$cache->[$key] = $a;
}
