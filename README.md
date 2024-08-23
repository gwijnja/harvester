# Harvester

Managed File Transfer library with detailed audit logging.

## Todo

- [ ] Perform memory profiling

## Install ##

```
go get -u github.com/gwijnja/harvester
```

## Documentation

AuditCopy!

## Examples

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

gzipper := gzip.Gzipper{}

archiver := local.Archiver{
    Transmit: "/path/to/archive/transmit",
    Archive:  "/path/to/archive",
    Regex:    "(\\d{4})(\\d{2})(\\d{2})",
    Format:   "$1/$2/$3",
}

writer := sftp.FileWriter{
    Connector: sftp.Connector{
        // Supports public key authentication.
        // See example further down below.
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
job.RunOnce()
```

## Local reader

## Local writer

## Download from FTP

Files in a 'to load' directory are listed. After downloading they are moved to the 'loaded' directory, unless you set DeleteAfterDownload to true.

```go
reader := ftp.FileReader{
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

Files are never uploaded directory to the target (ToLoad) directory, but always first to the Transmit directory. This is an intermediate step. After a succesful upload the file is moved to the ToLoad directory. This prevents growing files in the ToLoad directory.

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

## Upload to SFTP

## Writing to stdout

## Zip

## Unzip

## Gzip

## Gunzip

## Renaming

## Archiving

## Run once

## Run at interval

