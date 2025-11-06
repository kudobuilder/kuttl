.PHONY: test
test:
	$(MAKE) -C ../../ cli
	../../bin/kubectl-kuttl test --config /dev/null --start-control-plane --template-var var1=var1 --template-var var2=var2,var3=var3 suite1
