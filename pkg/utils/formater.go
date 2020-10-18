package utils

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"io"
	"sort"
	"text/tabwriter"
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

// FormatAsTable formats the given string entries as table where the header string defines the column titles.
func FormatAsTable(entries []string, header string) (string, error) {
	var errors *multierror.Error
	buffer := new(bytes.Buffer)

	writer := tabwriter.NewWriter(buffer, 0, 0, 1, ' ', tabwriter.Debug)
	_, e := fmt.Fprintf(writer, header)

	errors = multierror.Append(errors, e)

	errors = multierror.Append(errors, WriteSorted(entries, writer))
	errors = multierror.Append(errors, writer.Flush())

	if errors.Len() == 0 {
		return buffer.String(), nil
	}
	return "", errors.ErrorOrNil()
}
