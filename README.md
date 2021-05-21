# szip

All the benefits of tar with all the problems of zip




This is a simple library to allow opening zip files sequentially without
reading the [central directory][2] at the end. This means you sort of have to
know the file structure but not really if it is simple.

Built looking at the [wikipedia article][1] and the [go source code for
zip][3].

The interface is somehwat like the [tar reader interface][4], it has a:
- `func NewReader(r io.Reader) Reader`
- `func (r *Reader) Next() (zip.FileHeader, error)`
- `func (r Reader) Read(b []byte) (int, error)`

Additionally, it also has a `Close() error` just in case you like freeing
resources from the underlying [`flate`][5].

Originally designed for unzipping one-file steam manifests in the middle of the
download when the ending central directory hasn't been hit yet, and complements
go's http lazy downloading with that.


[1]: https://en.wikipedia.org/wiki/Zip_(file_format)
[2]: https://en.wikipedia.org/wiki/Zip_(file_format)#Central_directory_file_header
[3]: https://golang.org/src/archive/zip/
[4]: https://golang.org/pkg/archive/tar/#Reader
[5]: https://golang.org/pkg/compress/flate

### A note about permissions

Because zip record's permissions are stored in the central directory header
at the bottom of the file, this information cannot be acquired in-transit.

Because of this, after `Reader`'s EOF, `Reader.FileHeader` is populated with
the central directory header's information on a `Reader.CentralDirectory()`
call. This is otherwise unused during the extraction process, and is provided
to be used for reading additional metadata after extraction.


## Similar projects

### [stream-unzip](https://github.com/uktrade/stream-unzip)
- python
