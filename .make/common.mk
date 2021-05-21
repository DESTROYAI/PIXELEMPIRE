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
	## Do 