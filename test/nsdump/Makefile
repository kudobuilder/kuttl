# Verifies that when namespace deletion times out (in envtest there is no
# namespace controller, so a forced deletion always hangs), kuttl dumps the
# stuck namespace state and the objects remaining in it.

.PHONY: test
test:
	$(MAKE) -C ../../ cli
	rm -f kuttl-output.txt
	set +e; ../../bin/kubectl-kuttl test --config /dev/null --start-control-plane --timeout 10 --delete=all suite1 > kuttl-output.txt 2>&1; ret=$$?; set -e; \
	if [ $$ret -eq 0 ]; then echo "kuttl unexpectedly succeeded: namespace deletion must time out in envtest (no namespace controller)"; cat kuttl-output.txt; exit 1; fi
	grep -q "stuck namespace" kuttl-output.txt || { cat kuttl-output.txt; echo "FAIL: stuck namespace state was not dumped"; exit 1; }
	grep -q "ConfigMap/stuck-cm" kuttl-output.txt || { cat kuttl-output.txt; echo "FAIL: remaining objects were not dumped"; exit 1; }
