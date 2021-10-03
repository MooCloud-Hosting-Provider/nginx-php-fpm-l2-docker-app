
Work In Progress

## MooCLI

The following contains the possible commands to send as well as the range of inputs. Default values and data types are indicated by // after the option possibilities. If MooCLI receives input types other than specified, it will log an error and do nothing. If the parsing doesn't work for whatever reason, an appropriate error message is logged and no other change will happen.

### PHP


https://www.php.net/manual/en/ini.list.php
https://www.php.net/manual/en/faq.using.php#faq.using.shorthandbytes


```upload_max_filesize```: ```{1...64}``` // ```16``` (Integer) [Megabyte]


```post_max_size```: ```{1...64}``` // ```16``` (Integer) [Megabyte]


```memory_limit```: ```{1...512}``` // ```64``` (Integer) [Megabyte]


```allow_url_fopen```: ```{off,on}``` // ```off``` (String)


### Nginx


https://nginx.org/en/docs/http/ngx_http_fastcgi_module.html


```fastcgi_cache_location```: ```{drive,ramdisk}``` // ```drive``` (String)


```fastcgi_cache_profile```: ```{disabled,default}``` // ```default``` (String)


### Details

Some settings have dependencies.

For example in PHP, memory_limit must be
larger than post_max_size, which in turn must be
larger than upload_max_filesize.

If such an ordering is violated,
MooCLI will log an error without making any changes.


When specifying anything involving bytes,
it is usually a good idea to use powers of 2,
because any kind of hardware memory is aligned in powers of two. Why?

https://en.wikipedia.org/wiki/Data_structure_alignment

https://software.intel.com/content/www/us/en/develop/articles/coding-for-performance-data-alignment-and-structures.html

