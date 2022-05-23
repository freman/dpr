$(document).ready(function() {
	if (!window.rcmail) {
    		return
	}

	rcmail.addEventListener('init', function() {
		// Blow away the standard refresh timer
		rcmail.start_refresh = function() {console.log("noop");};
		clearInterval(rcmail._refresh)

		const evtSource = new EventSource(window.location.href.split('?')[0] + "sse/events");
		evtSource.onmessage = function(event) {
			rcmail.refresh();
		}
	})
})
