# Domainerator

Domainerator was my first Go application. That combines two wordlists (prefixes and suffixes) and a list of TLDs to form domain names and check their DNS status. It outputs a file with each combined domain name and the respective DNS status. 

## History

Not a long one, but...

I've developed it after getting tired of trying to find some good domain names available to be registered. 
In the beginning it was a slow Ruby script that still can be found as a rubygem with the same name, and now I've ported it to Go and made a lot of improvements in the method used to check for availability and speed.

## Benchmark

I was able to run a check with 2,288 domains in about 13 seconds using 100 goroutines in a MacBook Pro Late 2008!

## Instalation

To install domainerator you need:

1. Install Go
2. Setup environment vars for Go
3. Run `sudo go get github.com/hgfischer/domainerator`

## Updating

To update domainerator you need:

1. Run `sudo go get -u github.com/hgfischer/domainerator`
