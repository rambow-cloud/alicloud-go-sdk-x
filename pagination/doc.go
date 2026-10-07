// Package pagination provides a typed cursor engine for token and page-number
// adapters. Adapters decide termination from reviewed service metadata, including
// whether an empty page can still have a continuation token. Paginators are for
// one consumer and are not concurrency safe. Errors do not advance the cursor;
// repeated cursors terminate with ErrRepeatedCursor to prevent infinite loops.
package pagination
