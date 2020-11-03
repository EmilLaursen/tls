# tls 

Transliterates stdin to ASCII, with possibility to overwrite rules. It is a CLI wrapper around the wonderful [`github.com/alexsergivan/transliterator`](https://github.com/alexsergivan/transliterator), with concurrent processing if multiple files are supplied. Read from files or stdin.


## Usage
```bash
Transliterates stdin to ASCII, and also preserves æøå§, then outputs to stdout.
Supplied file arguments are processed concurrently. There is no concurrency bound
so do not supply a large amount of files, unless you want to use alot of file
descriptors.

		tls original/path/file1.txt

will place the transliterated file next to file1.txt at

		original/path/file1-transliterated.txt

Usage:
  tls file1.txt file2.txt [flags]

Examples:
cat file1.txt | tls
tls -v -o /my/awesome/dir file1.txt file2.txt file3.txt
	

Flags:
  -h, --help                help for tls
  -o, --output-dir string   custom output directory where transliterated files are placed
  -r, --rulesjson string    Custom transliteration rules. Json mapping unicode codepoints to strings
  -v, --verbose             show progress bars
```