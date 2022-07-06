package walk

import (
	"context"
	"gocloud.dev/blob"
	"io"
	_ "log"
	"strings"
	"sync"
)

func WalkBucket(ctx context.Context, opts *WalkOptions, bucket *blob.Bucket) error {

	error_ch := opts.ErrorChannel

	workers := opts.Workers

	throttle := make(chan bool, workers)

	for i := 0; i < workers; i++ {
		throttle <- true
	}

	wg := new(sync.WaitGroup)

	var walkFunc func(context.Context, *blob.Bucket, string) error

	walkFunc = func(ctx context.Context, bucket *blob.Bucket, prefix string) error {

		select {
		case <-ctx.Done():
			return nil
		default:
			// pass
		}

		iter := bucket.List(&blob.ListOptions{
			Delimiter: "/",
			Prefix:    prefix,
		})

		for {

			select {
			case <-ctx.Done():
				break
			default:
				// pass
			}

			obj, err := iter.Next(ctx)

			if err == io.EOF {
				break
			}

			if err != nil {

				e := &WalkError{
					Path:       prefix,
					LineNumber: 0,
					Err:        err,
				}

				error_ch <- e
				return nil
			}

			if obj.IsDir {

				err = walkFunc(ctx, bucket, obj.Key)

				if err != nil {

					e := &WalkError{
						Path:       obj.Key,
						LineNumber: 0,
						Err:        err,
					}

					error_ch <- e
				}

				continue
			}

			if obj.Size == 0 {
				continue
			}

			if opts.Filter != nil {

				if !opts.Filter(ctx, obj.Key) {
					continue
				}
			}

			// parse file of line-demilited records

			// trailing slashes confuse Go Cloud...

			path := strings.TrimRight(obj.Key, "/")

			go func(path string) {

				// log.Println("WAIT", path)
				<-throttle

				wg.Add(1)

				defer func() {
					// log.Println("CLOSE", path)
					wg.Done()
					throttle <- true
				}()

				fh, err := bucket.NewReader(ctx, path, nil)

				if err != nil {

					e := &WalkError{
						Path:       path,
						LineNumber: 0,
						Err:        err,
					}

					error_ch <- e
					return
				}

				defer fh.Close()

				opts.IsBzip = true

				if !strings.HasSuffix(path, ".bz2") {
					opts.IsBzip = false
				}

				ctx := context.WithValue(ctx, CONTEXT_PATH, path)

				WalkReader(ctx, opts, fh)

			}(path)
		}

		return nil
	}

	walkFunc(ctx, bucket, opts.URI)
	wg.Wait()

	return nil
}
