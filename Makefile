# ---------- basic ----------

OUT := bin

# ---------- duplicate ----------
DUPLICATE_NAME := duplicate
DUPLICATE_BIN := $(OUT)/$(DUPLICATE_NAME).exe
DUPLICATE_PKG := ./cmd/duplicate

DUPLICATE_SRC := cmd/duplicate/main.go
DUPLICATE_SRC += $(wildcard pkg/**/*.go)
DUPLICATE_SRC += $(wildcard internal/**/*.go)

# ---------- phony ----------

.PHONY: all build

all: build

# ---------- build ----------

build: $(DUPLICATE_BIN)

$(OUT):
	mkdir $(OUT)

$(DUPLICATE_BIN): $(OUT) $(DUPLICATE_SRC)
	go build -o $@ $(DUPLICATE_PKG)