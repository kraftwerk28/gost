package rxkbcommon

/*
#cgo LDFLAGS: -lxkbregistry
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <xkbcommon/xkbregistry.h>

void *init_ctx() {
	struct rxkb_context *ctx =
		rxkb_context_new(RXKB_CONTEXT_NO_DEFAULT_INCLUDES);
	rxkb_context_include_path_append_default(ctx);
	rxkb_context_parse_default_ruleset(ctx);
	return ctx;
}

void deinit_ctx(void *ctx) {
	rxkb_context_unref(ctx);
}

void *next_layout(void *ctx, void *layout) {
	return layout ? rxkb_layout_next(layout) : rxkb_layout_first(ctx);
}

const char *get_name(void *lay) {
	return rxkb_layout_get_name(lay);
}

const char *get_desc(void *lay) {
	return rxkb_layout_get_description(lay);
}
*/
import "C"

func GetLayoutShortNames() map[string]string {
	m := map[string]string{}
	ctx := C.init_ctx()
	defer C.deinit_ctx(ctx)
	layout := C.next_layout(ctx, nil)
	for layout != nil {
		name := C.get_name(layout)
		desc := C.get_desc(layout)
		m[C.GoString(desc)] = C.GoString(name)
		layout = C.next_layout(ctx, layout)
	}
	return m
}
