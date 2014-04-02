respond("!hello", function(msg) {
	msg.Reply("hi");
});

listen("001", function(msg) {
	noye.Join("#test");
});

respond("!reload", function(msg) {
	var scripts = core.scripts();
	log("%+v", scripts.Scripts);
	log("%t", _.contains(scripts.Scripts, "hello.js"));
});