.PHONY: test
test:
	$(MAKE) -C ../../ cli
	rm -f kuttl-report-step.xml.normalized kuttl-report-step.json.normalized
	rm -f kuttl-report-test.xml.normalized kuttl-report-test.json.normalized
	../../bin/kubectl-kuttl test --config /dev/null --report-name kuttl-report-step --report-granularity step --timeout 10 --report xml suite1 suite2 > kuttl-ouput-step-xml.txt 2>&1 || true # this is meant to fail
	../../bin/kubectl-kuttl test --config /dev/null --report-name kuttl-report-step --report-granularity step --timeout 10 --report json suite1 suite2 > kuttl-ouput-step-json.txt 2>&1 || true # this is meant to fail
	../../bin/kubectl-kuttl test --config /dev/null --report-name kuttl-report-test --report-granularity test --timeout 10 --report xml suite1 suite2 > kuttl-ouput-test-xml.txt 2>&1 || true # this is meant to fail
	../../bin/kubectl-kuttl test --config /dev/null --report-name kuttl-report-test --report-granularity test --timeout 10 --report json suite1 suite2 > kuttl-ouput-test-json.txt 2>&1 || true # this is meant to fail
	$(MAKE) kuttl-report-step.xml.normalized kuttl-report-step.json.normalized
	$(MAKE) kuttl-report-test.xml.normalized kuttl-report-test.json.normalized
	diff -u kuttl-report-step.xml.golden kuttl-report-step.xml.normalized
	diff -u kuttl-report-step.json.golden kuttl-report-step.json.normalized
	diff -u kuttl-report-test.xml.golden kuttl-report-test.xml.normalized
	diff -u kuttl-report-test.json.golden kuttl-report-test.json.normalized

.PHONY: update-golden
update-golden:
	cp kuttl-report-step.json.normalized kuttl-report-step.json.golden
	cp kuttl-report-step.xml.normalized kuttl-report-step.xml.golden
	cp kuttl-report-test.json.normalized kuttl-report-test.json.golden
	cp kuttl-report-test.xml.normalized kuttl-report-test.xml.golden

# The following targets replace all timestamps and durations with dummy values to make comparisons easy.

%.xml.normalized: %.xml
	sed -E -e 's/time="[^"]+"/time="1.0"/g; s/[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}[.][0-9]{6,10}[-+][0-9]{2}:[0-9]{2}/2000-01-01T00:00:00.00000000+00:00/g' < $< > $@

%.json.normalized: %.json
	sed -E -e 's/"time": *"[^"]+"/"time": "1.0"/g; s/[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}[.][0-9]{6,10}[-+][0-9]{2}:[0-9]{2}/2000-01-01T00:00:00.00000000+00:00/g' < $< > $@
