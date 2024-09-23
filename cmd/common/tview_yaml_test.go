package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToYAMLColored(t *testing.T) {
	yamlString := `
a: 1
b: 1.2
c: hej
d: false
e: true
f: null
g: []
h: {}
nested:
  a:
    b: 1
    c: hej
  d:
    e: null
    f: asdf
primitiveList: [1,2,3,null,{},[],hej,hej]
complexList:
  - key1: val1
    key2: val2
  - key1: val3
    key3: val4
  - {}
  - null
`

	expected := `[aqua]a[white]: [fuchsia]1
[aqua]b[white]: [fuchsia]1.2
[aqua]c[white]: [lime]hej
[aqua]d[white]: [fuchsia]false
[aqua]e[white]: [fuchsia]true
[aqua]f[white]: [white]null
[aqua]g[white]: [white][[white]]
[aqua]h[white]: [white]{}
[aqua]nested[white]:
  [aqua]a[white]:
    [aqua]b[white]: [fuchsia]1
    [aqua]c[white]: [lime]hej
  [aqua]d[white]:
    [aqua]e[white]: [white]null
    [aqua]f[white]: [lime]asdf
[aqua]primitiveList[white]: [white][[fuchsia]1[white], [fuchsia]2[white], [fuchsia]3[white], [white]null[white], [white]{}[white], [white][[white]][white], [lime]hej[white], [lime]hej[white]]
[aqua]complexList[white]:
[white]  - [aqua]key1[white]: [lime]val1
    [aqua]key2[white]: [lime]val2[white]
[white]  - [aqua]key1[white]: [lime]val3
    [aqua]key3[white]: [lime]val4[white]
[white]  - [white]{}[white]
[white]  - [white]null`

	res, err := ToYAMLColored(yamlString)
	require.NoError(t, err)
	require.Equal(t, expected, res)

}
