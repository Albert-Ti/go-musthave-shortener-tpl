package testreset

// generate:reset
type ResetableStruct struct {
	i   int
	str string
	s   []int
	m   map[string]string

	iP    *int
	strP  *string
	sP    *[]int
	mP    *map[string]int
	child *ResetableStruct

	ch chan string
}
