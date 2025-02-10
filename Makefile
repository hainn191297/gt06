build_lib_concox:
	gcc -c protocol/concox.c -o protocol/concox.o && ar rcs protocol/libconcox.a protocol/concox.o