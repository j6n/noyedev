respond("!reload (?P<script>.*?\.js)", function(msg, res) {
	var scripts = core.scripts()
	if (_.contains(scripts.Scripts, res.script)) {
		var err = core.manager.Reload(res.script);
		if (err) log("%+v", err);
	}
});