package main

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"cloud.google.com/go/storage"
	"github.com/alecthomas/kingpin/v2"
	"github.com/cockroachdb/errors"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var (
	flagWaste = kingpin.Flag("waste", "Run by waste memory.").Short('w').Bool()
)

func main() {
	kingpin.Parse()
	ctx := context.Background()
	client, err := storage.NewClient(ctx,
		option.WithoutAuthentication(),
		option.WithEndpoint("http://localhost:4443/storage/v1/"))
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	const bucket = "sample-bucket"
	if flagWaste != nil && *flagWaste {
		fmt.Println("Run by waste memory.")
		if err := makeBucketFilesZipWasteMemory(ctx, client, bucket); err != nil {
			fmt.Printf("%+v", err)
		}
		return
	}
	fmt.Println("Run by memory saving.")
	if err := makeBucketFilesZip(ctx, client, bucket); err != nil {
		fmt.Printf("%+v", err)
	}
}

func makeZipFileHeader(attrs *storage.ObjectAttrs) *zip.FileHeader {
	fh := &zip.FileHeader{
		Name:               attrs.Name,
		UncompressedSize64: uint64(attrs.Size),
		Modified:           attrs.Updated,
		Method:             zip.Deflate,
	}
	fh.SetMode(0755)
	return fh
}

//
// Memory saving
//

func makeBucketFilesZip(
	ctx context.Context,
	client *storage.Client,
	bucketName string,
) error {
	const zipName = "all.zip"
	bucket := client.Bucket(bucketName)
	it := bucket.Objects(ctx, &storage.Query{})

	writer := bucket.Object(zipName).NewWriter(ctx)
	defer writer.Close()
	zipWriter := zip.NewWriter(writer)
	defer zipWriter.Close()

	printMemoryStatsHeader()
	for {
		oattrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return errors.Wrap(err, "it.Next")
		}
		if err := readWriteOneFile(ctx, zipWriter, bucket, oattrs.Name); err != nil {
			return err
		}
		printMemoryStats(oattrs.Name)
	}
	printMemoryStats(zipName)
	return nil
}

func readWriteOneFile(
	ctx context.Context,
	zipWriter *zip.Writer,
	bucket *storage.BucketHandle,
	filename string,
) error {
	object := bucket.Object(filename)
	reader, err := object.NewReader(ctx)
	if err != nil {
		return errors.Wrap(err, "object.NewReader")
	}
	defer reader.Close()

	attrs, err := object.Attrs(ctx)
	if err != nil {
		return errors.Wrap(err, "object.Attrs")
	}
	header := makeZipFileHeader(attrs)
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return errors.Wrap(err, "zipWriter.CreateHeader")
	}
	_, err = io.Copy(writer, reader)
	if err != nil {
		return errors.Wrap(err, "io.Copy")
	}
	return nil
}

//
// Memory waste
//

func makeBucketFilesZipWasteMemory(
	ctx context.Context,
	client *storage.Client,
	bucketName string,
) error {
	const zipName = "all-waste.zip"
	bucket := client.Bucket(bucketName)
	var tmpFilePath string

	printMemoryStatsHeader()
	err := func() error {
		// create local file, waste memory and storage
		file, err := os.CreateTemp("", zipName)
		if err != nil {
			return errors.Wrap(err, "os.Open")
		}
		defer file.Close()
		tmpFilePath = file.Name()
		zipWriter := zip.NewWriter(file)
		defer zipWriter.Close()

		it := bucket.Objects(ctx, &storage.Query{})
		for {
			oattrs, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return errors.Wrap(err, "it.Next")
			}
			if err := readWriteOneFileWasteMemory(ctx, zipWriter, bucket, oattrs.Name); err != nil {
				return err
			}
			printMemoryStats(oattrs.Name)
		}
		return nil
	}()
	if err != nil {
		return err
	}

	reader, err := os.Open(tmpFilePath)
	if err != nil {
		return errors.Wrap(err, "os.Open")
	}
	defer func() {
		reader.Close()
		os.Remove(tmpFilePath)
	}()
	// ReadAll() waste memory
	readerBytes, err := io.ReadAll(reader)

	printMemoryStats(zipName)

	writer := bucket.Object(zipName).NewWriter(ctx)
	_, err = writer.Write(readerBytes)
	writer.Close()
	if err != nil {
		return errors.Wrap(err, "writer.Write")
	}

	return nil
}

func readWriteOneFileWasteMemory(
	ctx context.Context,
	zipWriter *zip.Writer,
	bucket *storage.BucketHandle,
	filename string,
) error {
	object := bucket.Object(filename)
	reader, err := object.NewReader(ctx)
	if err != nil {
		return errors.Wrap(err, "object.NewReader")
	}
	defer reader.Close()

	attrs, err := object.Attrs(ctx)
	if err != nil {
		return errors.Wrap(err, "object.Attrs")
	}
	header := makeZipFileHeader(attrs)
	w, err := zipWriter.CreateHeader(header)
	if err != nil {
		return errors.Wrap(err, "zipWriter.CreateHeader")
	}
	// ReadAll() waste memory
	readerBytes, err := io.ReadAll(reader)
	if err != nil {
		return errors.Wrap(err, "io.ReadAll")
	}
	if _, err := w.Write(readerBytes); err != nil {
		return errors.Wrap(err, "w.Write")
	}
	return nil
}
