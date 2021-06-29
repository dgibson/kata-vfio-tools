CRICTL = sudo crictl

run-%:
	$(CRICTL) start $$($(CRICTL) create $$($(CRICTL) runp pod.json) container-$*.json pod.json)

sh-%: run-%
	$(CRICTL) exec -it $$($(CRICTL) ps -lq) /bin/sh

clean:
	rm -f *~

cleaner: clean
	for c in $$($(CRICTL) ps -aq); do $(CRICTL) stop $$c && $(CRICTL) rm $$c; done
	for p in $$($(CRICTL) pods -q); do $(CRICTL) stopp $$p && $(CRICTL) rmp $$p; done
