package database

type PaginatedRecord[T any] struct {
	Data       []T // Generic array of data
	PageNumber int // Current page number
	PageSize   int // Number of records per page
	Total      int // Total number of records
}
