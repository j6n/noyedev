respond("!hello", function(msg) {
	msg.Reply("hi");
});

listen("001", function(msg) {
	noye.Join("#test");
})