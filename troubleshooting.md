# Installing cfssl

Error seen:

```
go build github.com/mattn/go-sqlite3
		# github.com/mattn/go-sqlite3
		sqlite3-binding.c: In function ‘sqlite3SelectNew’:
		sqlite3-binding.c:128049:10: warning: function may return address of local variable [-Wreturn-local-addr]
		128049 |   return pNew;
		       |          ^~~~
		sqlite3-binding.c:128009:10: note: declared here
		128009 |   Select standin;
		       |          ^~~~~~~

```

Solution shown here:
https://github.com/mattn/go-sqlite3/issues/803

```
export CGO_CFLAGS="-g -O2 -Wno-return-local-addr" go install github.com/cloudflare/cfssl/cmd/cfssl@v1.4.1
```

# Authorize with Access Control Lists, Chapter 5

The author has us create a policy.csv and a model.conf. Format matters!!! We had the wrong format.

the original csv file I had without formatting and it caused bugs:
The kindle cloud e-reader wasn't displaying this snippet correctly.
```
p,root,*,produce
p,root,*,consume
```

the fixed format:
I pulled this from the repo the author hosts. 
The snippet is, however, displayed correctly in the kindle app
on my tablet.
```
p, root, *, produce
p, root, *, consume
```
