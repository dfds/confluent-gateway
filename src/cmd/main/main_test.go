package main

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMust_Ok(t *testing.T) {
	got := Must("dummy", nil)

	assert.Equal(t, "dummy", got)
}

func TestMust_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Missing panic")
		}
	}()

	_ = Must("dummy", errors.New("some-error"))
}
