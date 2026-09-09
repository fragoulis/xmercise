#!/bin/sh
# Exercise one success case and one validation failure for every API operation.
set -eu

base_url=${BASE_URL:-http://localhost:8080}
tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT HUP INT TERM

request() {
	request_name=$1
	expected_status=$2
	shift 2
	if [ -n "${TOKEN:-}" ]; then
		set -- --header "Authorization: Bearer $TOKEN" "$@"
	fi

	status=$(curl --silent --show-error \
		--output "$tmpdir/$request_name.json" \
		--write-out '%{http_code}' \
		"$@")

	if [ "$status" != "$expected_status" ]; then
		echo "$request_name: got HTTP $status, want $expected_status" >&2
		cat "$tmpdir/$request_name.json" >&2
		exit 1
	fi

	echo "$request_name: HTTP $status"
}

name=$(printf 'c%s%s' "$(date +%s)" "$$" | cut -c1-15)

request create-validation 400 \
	--request POST \
	--header 'Content-Type: application/json' \
	--data '{"name":"","employees_count":-1,"registered":true,"type":"invalid"}' \
	"$base_url/v1/companies"

request create 201 \
	--request POST \
	--header 'Content-Type: application/json' \
	--data "{\"name\":\"$name\",\"description\":\"Initial description\",\"employees_count\":1,\"registered\":true,\"type\":\"Corporations\"}" \
	"$base_url/v1/companies"

company_id=$(sed -n 's/.*"id":"\([^"]*\)".*/\1/p' "$tmpdir/create.json")
if [ -z "$company_id" ]; then
	echo 'create: response does not contain an ID' >&2
	cat "$tmpdir/create.json" >&2
	exit 1
fi

request get-validation 400 "$base_url/v1/companies/not-a-uuid"
request get 200 "$base_url/v1/companies/$company_id"

request patch-validation 400 \
	--request PATCH \
	--header 'Content-Type: application/merge-patch+json' \
	--data '{"employees_count":-1}' \
	"$base_url/v1/companies/$company_id"

request patch 200 \
	--request PATCH \
	--header 'Content-Type: application/merge-patch+json' \
	--data '{"employees_count":2,"description":null}' \
	"$base_url/v1/companies/$company_id"

request delete-validation 400 --request DELETE "$base_url/v1/companies/not-a-uuid"
request delete 204 --request DELETE "$base_url/v1/companies/$company_id"
