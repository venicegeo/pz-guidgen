#!/bin/sh

gometalinter \
--deadline=60s \
--concurrency=6 \
--vendor \
--exclude="exported (var)|(method)|(const)|(type)|(function) [A-Za-z\.0-9]* should have comment" \
--exclude="comment on exported function [A-Za-z\.0-9]* should be of the form" \
--exclude="Api.* should be .*API" \
--exclude="Http.* should be .*HTTP" \
--exclude="Id.* should be .*ID" \
--exclude="Json.* should be .*JSON" \
--exclude="Url.* should be .*URL" \
./...

