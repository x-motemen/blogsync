package main

import "time"

const clockSkewAllowance = 3 * time.Second

func newerWithAllowance(local, remote time.Time) bool {
	return !local.Add(clockSkewAllowance).Before(remote)
}

func nowAfterWithAllowance(t time.Time) bool {
	return !time.Now().Add(clockSkewAllowance).Before(t)
}
