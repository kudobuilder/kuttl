.PHONY: test
test:
	$(MAKE) -C ../../ cli
	rm -f kuttl-report-step.xml.normalized kuttl-report-step.json.normalized
	../../bin/kubectl-kuttl test --config /dev/null --start-control-plane --report-name kuttl-report-step --timeout 10 --report xml suite1 suite2 > kuttl-ouput-step-xml.txt 2>&1 || true # this is meant to fail
	if [ ! -e kuttl-report-step.xml ]; then cat kuttl-output-step-xml.txt; exit 1; fi
	../../bin/kubectl-kuttl test --config /dev/null --start-control-plane --report-name kuttl-report-step --timeout 10 --report json suite1 suite2 > kuttl-ouput-step-json.txt 2>&1 || true # this is meant to fail
	if [ ! -e kuttl-report-step.json ]; then cat kuttl-output-step-json.txt; exit 1; fi
	$(MAKE) kuttl-report-step.xml.normalized kuttl-report-step.json.normalized
	diff -u kuttl-report-step.xml.golden kuttl-report-step.xml.normalized
	diff -u kuttl-report-step.json.golden kuttl-report-step.json.normalized

.PHONY: update-golden
update-golden:
	cp kuttl-report-step.json.normalized kuttl-report-step.json.golden
	cp kuttl-report-step.xml.normalized kuttl-report-step.xml.golden

# The following targets replace all timestamps and durations with dummy values to make comparisons easy.

%.xml.normalized: %.xml
	sed -E -e 's/time="[^"]+"/time="1.0"/g; s/[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}[.][0-9]{6,10}(Z|[-+][0-9]{2}:[0-9]{2})/2000-01-01T00:00:00.00000000+00:00/g' < $< > $@

%.json.normalized: %.json
	sed -E -e 's/"time": *"[^"]+"/"time": "1.0"/g; s/[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}[.][0-9]{6,10}(Z|[-+][0-9]{2}:[0-9]{2})/2000-01-01T00:00:00.00000000+00:00/g' < $< > $@
