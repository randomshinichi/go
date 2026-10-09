use strict;
my (%value, %is_errno, %is_signal);
open(my $nf, '<', 'zerrors_names.txt') or die $!;
my @names = <$nf>; chomp @names; close $nf;
open(my $vf, '<', 'constants.txt') or die $!;
while (<$vf>) { chomp; my ($name, $value) = split /=/, $_, 2; $value{$name} = $value; }
close $vf;
for my $spec ([errno => 'zerrors_errno_names.txt', \%is_errno], [signal => 'zerrors_signal_names.txt', \%is_signal]) {
	my ($kind, $path, $set) = @$spec;
	open(my $f, '<', $path) or die $!;
	while (<$f>) { chomp; $set->{$_} = 1; }
	close $f;
}
sub hexval { my ($v) = @_; return $v < 0 ? sprintf('-0x%x', -$v) : sprintf('0x%x', $v); }
sub quoted { my ($s) = @_; $s =~ s/\\/\\\\/g; $s =~ s/"/\\"/g; return '"'.$s.'"'; }
print "// Code generated on pmg5 from Leopard ppc32 headers; DO NOT EDIT.\n";
print "// Values: gcc -arch ppc C probes; tables: strerror/strsignal on Darwin 9.8.0.\n\npackage syscall\n\n";
for my $group (0,1,2) {
	print "const (\n";
	for my $name (@names) {
		next unless exists $value{$name};
		next if $group == 0 && ($is_errno{$name} || $is_signal{$name});
		next if $group == 1 && !$is_errno{$name};
		next if $group == 2 && !$is_signal{$name};
		my $expr = hexval($value{$name});
		$expr = "Errno($expr)" if $group == 1;
		$expr = "Signal($expr)" if $group == 2;
		print "\t$name = $expr\n";
	}
	print ")\n\n";
}
open(my $sf, '<', 'strings.txt') or die $!;
my $table;
while (<$sf>) {
	chomp;
	if ($_ eq 'ERRORS') { $table = 'errors'; print "// Error table\nvar errors = [...]string{\n"; next; }
	if ($_ eq 'SIGNALS') { $table = 'signals'; print "}\n\n// Signal table\nvar signals = [...]string{\n"; next; }
	my ($n, $str) = split /\t/, $_, 2;
	print "\t$n: ".quoted($str).",\n";
}
close $sf;
print "}\n";
