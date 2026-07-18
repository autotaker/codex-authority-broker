CC?=cc
CFLAGS?=-O2 -fPIC -Wall -Wextra -Werror -fstack-protector-strong
.PHONY: all check
all: pam_codex_lease.so
pam_codex_lease.so: pam_codex_lease.c
	$(CC) $(CFLAGS) -shared -Wl,-z,relro,-z,now -o $@ $< -lpam
check:
	@if test -f /usr/include/security/pam_modules.h; then $(MAKE) all; else echo 'SKIP: PAM headers unavailable'; fi
