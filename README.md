Domainerator
============

Domainerator is an application written in Go that combines a wordlist and a list of TLDs to form domain names and check if they are not registered domains.

It was ported from a Ruby version made a few years ago that still can be found as a ruby gem, but it is too slow. This version made in Go is blazing fast and one of the reasons is that it can do checks in parallel. I was able to run a process that checked 110,695 domains in about 31 minutes using 200 goroutines in a MacBook Pro Late 2008 and returned 108,237 available domains.

Currently the check is done by resolving the domain apex. However, this method is not reliable since some registars answers with some default IP address even for unregistered domains (eg. .ws TLD). I have plans to add a whois check to it soon.
