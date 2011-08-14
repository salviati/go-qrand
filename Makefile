include $(GOROOT)/src/Make.inc

all: install

DIRS=qrand


clean.dirs: $(addsuffix .clean, $(DIRS))
install.dirs: $(addsuffix .install, $(DIRS))
nuke.dirs: $(addsuffix .nuke, $(DIRS))

%.clean:
	+cd $* && gomake clean

%.install:
	+cd $* && gomake install

%.nuke:
	+cd $* && gomake nuke

clean: clean.dirs
install: install.dirs
nuke: nuke.dirs
