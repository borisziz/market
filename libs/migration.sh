#!/bin/sh
goose -dir $2 postgres $1 status
goose -dir $2 postgres $1 up