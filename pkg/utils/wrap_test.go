package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_WordWrap(t *testing.T) {
	s := `Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book.
It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum.

Here are some bullet points:
  - point 1: hej
  - point 2: hej again
  - point 3: word word word word word word word word word word word word word word word word word word word word word`

	expected := `Lorem  Ipsum  is  simply  dummy  text  of  the  printing and
typesetting  industry.  Lorem  Ipsum has been the industry's
standard  dummy  text  ever since the 1500s, when an unknown
printer  took  a  galley  of type and scrambled it to make a
type specimen book.
It  has  survived not only five centuries, but also the leap
into electronic typesetting, remaining essentially
unchanged.  It was popularised in the 1960s with the release
of Letraset sheets containing Lorem Ipsum passages, and more
recently   with   desktop  publishing  software  like  Aldus
PageMaker including versions of Lorem Ipsum.

Here are some bullet points:
  - point 1: hej
  - point 2: hej again
  -  point  3:  word word word word word word word word word
word  word word word word word word word word word word word
`

	wrapped := WordWrap(s, 60, "")
	require.Equal(t, expected, wrapped)
}
