# Harvester MFT

Managed File Transfer library with detailed audit logging, including source and destination SHA-1 file hashes. Source, destination and intermediate steps are linked though the *io.Reader* interface, so files do not need to be stored locally as intermediate steps.

## Installation ##

```
go get -u github.com/gwijnja/harvester
```

## Overview

To use this library you need to prepare an MFT job. Every job contains at least a reader and a writer, which is one of the following:

* Local directory
* FTP source/destination
* SFTP source/destination
* Stdout (for testing)

Next, you can add intermediate steps. Currently the following processes are supported:

* Zip/unzip
* Gzip/gunzip
* Renaming
* Archiving

Files are never written directly to a destination directory, but always via an intermediate directory, to prevent growing files in the destination directory.

It is **important** that the intermediate directory is on the same filesystem as the destination directory. Otherwise there will still be a growing file when it is moved from the intermediate to the destination directory.

## Examples

Let's look at a complete example first, to give you an idea of what it looks like.

In this example files are taken from a local directory, renamed, gzipped, archived, and finally uploaded to an SFTP location. The reader and writer are mandatory. You can insert optional processing steps in between.

```go
reader := local.FileReader{
    ToLoad:              "/home/user/mft/toload",
    Loaded:              "/home/user/mft/loaded",
    FollowSymlinks:      true,
    DeleteAfterDownload: false,
    Regex:               "\\.csv$",
}

renamer := harvester.Renamer{
	Regex:  "(\\d{4})-(\\d{2})-(\\d{2})",
	Format: "output_$1-$2-$3.csv",
}

gzipper := gzip.Compressor{}

archiver := local.Archiver{
    Transmit: "/path/to/archive/transmit",
    Archive:  "/path/to/archive",
    Regex:    "(\\d{4})(\\d{2})(\\d{2})",
    Format:   "$1/$2/$3",
}

writer := sftp.Uploader{
    Connector: sftp.Connector{
        // Public key authentication is supported,
        // see examples further down below.
        Host:     "sftp.example.com",
        Port:     2222,
        Username: "itsme",
        Password: "s3cr3t",
    },
    Transmit: "/remote/mft/transmit",
    ToLoad:   "/remote/mft/toload",
}

job := harvester.NewJob(&reader, &writer)
job.Insert(&renamer)
job.Insert(&gzipper)
job.Insert(&archiver)

interval := 5 * time.Minute
job.Run(interval) // or run once using RunOnce()
```

## Local reader

Files can originate from three locations:

* A local directory
* An FTP source
* An SFTP source

The example below is for local directories. The *FileReader* lists the files in `ToLoad` and filters them according to the `Regex`. If the regular expression is left empty, all files match. After downloading they are either moved to the `Loaded` directory, other deleted, depending on the `DeleteAfterDownload` variable. If `DeleteAfterDownload` is true, then the `Loaded` field can be left empty.

Refer to the [RE2 syntax](https://github.com/google/re2/wiki/Syntax) for help with writing regular expressions.

`MaxFiles` limits the number of files per job run. If you set it to 0, then there is no limit.

```go
reader := local.FileReader{
    ToLoad:              "/path/to/toload",
    Loaded:              "/path/to/loaded",
    FollowSymlinks:      true,
    Regex:               "\\.csv$",
    MaxFiles:            10,
    DeleteAfterDownload: false,
}
```

## Local writer

Files can be written to three locations:

* A local directory
* An FTP destination
* An SFTP destination

```go
writer := local.FileWriter{
    Transmit: "/path/to/transmit",
    ToLoad:   "/path/to/toload",
}
```

Files are first written to `Transmit` and then moved to `ToLoad`.

## Download from FTP

This uses the https://github.com/jlaffaye/ftp client. The downloader lists the files in `ToLoad` and filters them according to the `Regex`. If the regular expression is left empty, all files match. After downloading they are either moved to the `Loaded` directory, other deleted, depending on the `DeleteAfterDownload` variable. If `DeleteAfterDownload` is true, then the `Loaded` field can be left empty.

```go
reader := ftp.Downloader{
    Connector: ftp.Connector{
        Host:     "ftp.example.com",
        Port:     21,
        Username: "itsme",
        Password: "s3cr3t",
    },
    ToLoad:              "/mft/toload",
    Loaded:              "/mft/loaded",
    Regex:               ".*\\.txt",
    DeleteAfterDownload: false,
}
```

## Upload to FTP

Files are upload to the intermediate `Transmit` directory, and then moved to the final `ToLoad` directory. This is to prevent growing files to appear in the `ToLoad` directory.

```go
writer := ftp.FileWriter{
    Connector: ftp.Connector{
        Host:     "ftp.example.com",
        Port:     21,
        Username: "itsme",
        Password: "s3cr3t",
    },
    Transmit: "/mft/transmit",
    ToLoad:   "/mft/toload",
}
```

## Download from SFTP

This connector uses Go's crypto/ssh package and https://github.com/pkg/sftp for the SFTP client. To simply connect using username and password:

```go
reader := sftp.Downloader{
    Connector: sftp.Connector{
        Host:     "localhost",
        Port:     22,
        Username: "itsme",
        Password: "s3cr3t",
    },
    ToLoad:              "/path/to/toload",
    Loaded:              "/path/to/loaded",
    Regex:               "^orders-\\d+\\.json$",
    MaxFiles:            10,
    DeleteAfterDownload: true,
}
```

The paths in this example are absolute, but relative paths should work just as well. Note that there is no 'root' directory, so the client does not *cd* into another directory after logging in. All actions are performed relative to the directory it lands in after logging in.

The SFTP connector supports more options. This is the full *Connector* struct:

```go
type Connector struct {
	Host                  string
	Port                  int
	FailIfHostKeyChanged  bool
	FailIfHostKeyNotFound bool
	Username              string
	Password              string
	PrivateKeyFile        string
	Passphrase            string
}
```

By default the connection still works if the host key has changed, or if a host key is not present in the *known_hosts* file. This behaviour can be changed using the `FailIfHostKeyChanged` and `FailIfHostKeyNotFound` booleans.

To connect using public key authentication, specify the path to the private key with the `PrivateKeyFile`. The public key does not need to be specified, because the public keycan be derived from the private key. Therefore the SSH library only needs the private key.

If the private key is protected by a password (usually called a *passphrase* in the context of private keys), you can set it in the `Passphrase` field. Writing your private key passphrase in a configuration file or code comes with a risk, but I assume you know what you are doing.

Unlike OpenSSH, Go's SSH client accepts DSA keys out of the box. (See [here](https://cs.opensource.google/go/x/crypto/+/refs/tags/v0.26.0:ssh/handshake.go;l=153) and [here](https://cs.opensource.google/go/x/crypto/+/refs/tags/v0.26.0:ssh/common.go;l=73)). I will look into manually enabling other host key algorithms, key exchange algorithms, ciphers etc later.

## Upload to SFTP

The *Connector* part of the uploader is the same as described above with the SFTP downloader. Other than that there is just the `Transmit` and `ToLoad` fields that were discussed before.

```go
writer := sftp.Uploader{
    Connector: sftp.Connector{
        Host:     "localhost",
        Port:     22,
        Username: "itsme",
        Password: "s3cr3t",
    },
    Transmit: "/path/to/transmit",
    ToLoad:   "/path/to/toload",
}
```

## Writing to stdout

There is a stdout writer, which you can use for testing. It has no options:

```go
writer := stdout.Printer{}
```

## Zip

Just like the gzip compressor, the zip compressor has no options (yet):

```go
compressor := zip.Compressor{}
```

One important difference with the gzip compressor is how the file is renamed. When compressing a file with *gzip*, the `.gz` extension is **added**. When compressing using *zip* the extension is **replaced** with `.zip`.

Example: `foo.txt` compressed with *gzip* becomes `foo.txt.gz`. But `foo.txt` compressed with *zip* becomes `foo.zip`. Since this is the standard behaviour on the command line, I adapted it in this library.

The input filename is stored inside the archive. This is unlike gzip, which only compresses the bytestream but does not store filenames inside. (That's what tar is for, as in *.tar.gz*.)

## Unzip

Unzipping is just as simple:

```go
decompressor := zip.Decompressor{}
```

The output filename is not derived from the input filename, but extracted from the file entry in the archive.

The decompressor expects exactly one file in the archive. If there are multiple files, or if the file is inside a directory in the archive, then the input file is rejected.

## Gzip

Files are compressed inline using Go's `compress/gzip` package. The `.gz` extension will be added to filenames. Insert a gzip compressor into the chain using an empty Compressor struct. The struct has no further options (yet).

```go
compressor := gzip.Compressor{}
job.Insert(&compressor)
```

## Gunzip

Works the same as the gzip compressor. If the input file has a `.gz` extension, it is removed. The struct has no further options.

```go
decompressor := gzip.Decompressor{}
job.Insert(&decompressor)
```

## Renaming

Files can be renamed at any point in the chain, even multiple times, for example before and after compressing a file.

For example, let's say your input filenames contain dates and times: `report_20240825_212430.csv`

But you don't like underscores and you want to replace them with dashes, so it becomes `report-20240825-212430.csv`

Then you want to zip the file, and finally remove the timestamp from the zip filename so it becomes `report-20240825.zip`

So there will be two renamers and one zipper:

```go
renamer1 := harvester.Renamer{
	Regex: "^report_(\\d{8})_(\\d{6})\\.csv$",
	Format: "report-$1-$2.csv",
}

zipper := zip.Compressor{}

renamer2 := harvester.Renamer{
    Regex: "^report-(\\d{8})-\\d{6}\\.zip$",
    Format: "report-$1.zip",
}

job := harvester.NewJob(&reader, &writer) // r/w creation omitted
job.Insert(&renamer1)
job.Insert(&zipper)
job.Insert(&renamer2)
job.RunOnce()
```

I'm showing full regular expressions, but of course you can make them smaller, like `Regex: "(\\d{8})"` which just matches the 8 digits. Do remember to use double backslashes for proper escaping.

## Archiving

The archiver can be used to copy files to an archive, in the middle of a chain, and automatically create directories for year/month/day or whatever format you want. This is customizable through the `Regex` and `Format` fields.

For example, let's say you have files in the format of `orders-2024-08-25.csv` and you want them to be copied to an archive during a file transfer

```go
reader := ftp.Downloader{bla..}
writer := sftp.Uploader{bla..}

archiver := local.Archiver{
	Transmit: "/mnt/archive/transmit",
	Archive:  "/mnt/archive/orders",
	Regex:    "(\\d{4})-(\\d{2})-(\\d{2})", // match: yyyy-mm-dd
	Format:   "$1/$2/$3",                   // directory: yyyy/mm/dd
}
```

The regular expression matches the date in the filename, and each group of digits is captured separately because of the parentheses around them. So yyyy becomes $1, mm becomes $2 and dd becomes $3.

Then in the `Format` field you write the directory pattern you want to use in the archive. In this case it's `$1/$2/$3`, so the archiver would create directory `/mnt/archive/orders/2024/08/25` and place the file in it.

To prevent growing files in the archive, the file is first created in the `Transmit` directory, and then moved to the final directory.

It can currently not be used as a writer! It **must** be inserted into the middle.

## Run once

If you want to run the job only once, for testing or maybe because you always run the job from a cron table, then call it like the big example at the top:

```go
job.RunOnce()
```

## Run at interval

To automatically run the job at a specified interval, start the job with a *time.Duration* as an argument, which will be used for *time.Sleep()*:

```go
interval := 10 * time.Minute
job.Run(interval)
```

Please note that the job does not yet run with cancel context, and does not yet listen to signals, so there is no graceful shutdown, and therefore should not yet be used as a service. That is work in progress.

## Logging

The package is currently outputting a lot of logging, using [Go's slog](https://go.dev/blog/slog) package. The slog package supports changing the default logging, so you configure the output format prior to starting a harvester job. For example, you can output in JSON format, and enable the debug level:

```go
func main() {
	handlerOpts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, handlerOpts))
	slog.SetDefault(logger)

    // And then use the harvester library
}
```

I plan on adding the possibility to extend logging functionality; you should be able to pass a logger to the job, and every job run and file transfer should have a unique identifier.
