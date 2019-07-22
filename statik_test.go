package main

import "testing"

func TestToSymbolSafe(t *testing.T) {
	testCase := [][]string{
		{"abc", "Abc"},
		{"_abc", "Abc"},
		{"3abc", "Abc"},
		{"abc3", "Abc3"},
		{"/abc", "Abc"},
		{"abc abc", "AbcAbc"},
	}
	for i, test := range testCase {
		got := toSymbolSafe(test[0])
		wont := test[1]
		if got != wont {
			t.Errorf("#%02d toSymbolSafe(%s) => %s != %s", i, test[0], got, wont)
		}
	}
}
