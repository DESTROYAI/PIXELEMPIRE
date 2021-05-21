## Default repository domain name
ifndef GIT_DOMAIN
	override GIT_DOMAIN=github.com
endif

## Set if defined (alias variable for ease of use)
ifdef branch
	override REPO_BRANCH=$(branch)
	export REPO_BRANCH
endif

## Do we have git available?
HAS_GIT := $(shell command -v git 2> /dev/null)

ifdef HAS_GIT
	## Do we have a repo?
	HAS_REPO := $(shell git rev-parse --is-inside-work-tree 2> /dev/null)
	ifdef HAS_REPO
		## Automatically detect the repo owner and repo name (for local use with Git)
		REPO_NAME=$(shell basename "$(shell git rev-parse --show-toplevel 2> /dev/null)")
		OWNER=$(shell git config --get remote.origin.url | sed 's/git@$(GIT_DOMAIN)://g' | sed 's/\/$(REPO_NAME).git//g')
		REPO_OWNER=$(shell echo $(OWNER) | tr A-Z a-z)
		VERSION_SHORT=$(shell git describe --tags --always --abbrev=0)
		export REPO_NAME, REPO_OWNER, VERSION_SHORT
	endif
