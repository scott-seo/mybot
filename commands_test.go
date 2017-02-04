package main

import "testing"

func TestPut(t *testing.T) {
	setdebug("off")
	put("default yahoo http://www.yahoo.com")
	get("default yahoo | healthcheck | hello")
	setdebug("off")
}

func TestRemp(t *testing.T) {
	setdebug("off")
	put("default yahoo http://www.yahoo.com")
	get("default yahoo | healthcheck | echo")
	setdebug("off")
}
