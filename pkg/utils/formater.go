package utils

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"io"
	"sort"
)

// WriteSorted writes all elements from entries sorted into the given writer.
func WriteSorted(entries []string, writer io.Writer) error {
	var errors *multierror.Error
	sort.Strings(entries)
	for _, v := range entries {
		_, e := fmt.Fprint(writer, v)
		errors = multierror.Append(errors, e)
	}
	return errors.ErrorOrNil()
}
