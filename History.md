
1.7.0 / 2012-10-20 
==================

  * Added support for protocol
  * Changed list of default public suffix and timeouts
  * Improved checks
  * removed unused import
  * Bug fix in retries

1.6.9 / 2012-10-17 
==================

  * Check empty domain list

1.6.8 / 2012-10-17 
==================

  * Add filtering of restricted domains
  * Added minLength size limit
  * Setup of Travis-CI

1.6.7 / 2012-10-14 
==================

  * Added Travis-CI conf
  * Deduplicated domain list

1.6.6 / 2012-10-13 
==================

  * Fixed bug with -avail cmd option

1.6.5 / 2012-10-13 
==================

  * Included option to output only available domains
  * Changed default PS list
  * Fixed domain hack generation or small words

1.6.4 / 2012-10-13 
==================

  * Refactoring to include tests
  * Extraced wordlist functions to another package and tested all them

1.6.3 / 2012-10-12 
==================

  * Updated README and default list os public suffixes
  * Bigger default TLDs and bug fix in domain hacks

1.6.2 / 2012-10-12 
==================

  * Fixed bug in loop logic

1.6.1 / 2012-10-11 
==================

  * Added max domain length option

1.6.0 / 2012-10-11 
==================

  * Fixed TLD/Public Suffix naming confusion
  * Bigger default DNS server list
  * Domain hacks support
  * Flag to skip UTF-8 domains including UTF-8 public suffixes
  * Flag to include all TLDs in the public domain suffix list into the check

1.5.1 / 2012-10-11 
==================

  * Fixed mess/confusion with workspace/packages/github

1.5.0 / 2012-10-11 
==================

  * Better README
  * Embedded TLD list 
  * Improvements in DNS check
  * Output file now have domain and DNS answer status (NXDOMAIN == available)

1.4.0 / 2012-10-10 
==================

  * Now querying for NS instead of apex A
  * Added DNS option
  * Added github.com/miekg/dns/ as DNS client

1.3.1 / 2012-10-10 
==================

  * Added call to remove duplicated from domain list

1.3.0 / 2012-10-07 
==================

  * Added support for separate word lists for prefixes and suffixes
  * Added more items to git ignore list
  * Added install instructions

1.2.0 / 2012-10-07 
==================

  * Added support for checking suffixes and TLDs
  * Added output to file and better feedback to user
  * Added BSD license
  * Added history file

1.1.0 / 2012-10-06 
==================

  * Added new options

1.0.1 / 2012-10-06 
==================

  * small fixes

1.0.0 / 2012-10-06 
==================

  * first functional version
