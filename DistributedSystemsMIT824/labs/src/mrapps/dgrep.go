package main

//
// a distributed-grep application "plugin" for MapReduce.
//
// go build -buildmode=plugin dgrep.go
//

import (
	"strings"

	"6.824/mr"
)

const (
	Needle = "ADVENTURE"
)

//
// The map function is called once for each file of input. The first
// argument is the name of the input file, and the second is the
// file's complete contents. You should ignore the input file name,
// and look only at the contents argument. The return value is a slice
// of key/value pairs.
//
func Map(filename string, contents string) []mr.KeyValue {
	// split contents into an array of lines
	lines := strings.Split(contents, "\n")

	kva := []mr.KeyValue{}
	for _, line := range lines {
		// Case senstiive match
		// To get case insensitive match, simply compare lowercase the strings
		if strings.Contains(line, Needle) {
			kv := mr.KeyValue{Needle, line}
			kva = append(kva, kv)
		}
	}
	return kva
}

//
// The reduce function is called once for each key generated by the
// map tasks, with a list of all the values created for that key by
// any map task.
//
func Reduce(key string, values []string) string {
	// list all the lines having the key
	return strings.Join(values, "\n")
}
