noye = {
  "bot":  _noye_bot,
  "auth": [],
};

core = {
  "manager": _core_manager,
  "scripts": _core_scripts,

  "load": _core_storage_load,
  "save": _core_storage_save,
};

share = {
  "init":   _share_init,
  "update": _share_update,

  "sub":   _share_sub,
  "unsub": _share_unsub,
};

http = {
  "get": _http_get,
};

share.init("auth", function(data) {
  noye.auth = JSON.parse(data);
});