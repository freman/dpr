<?php
// Helper to return the username to dpr in exchange for borrowed session credentials
class sse extends rcube_plugin {
	function init() {
		$this->include_script('sse.js');
		$this->register_action('plugin.sse', array($this, 'request_handler'));
	}

	function request_handler() {
		$rcmail = rcmail::get_instance();

	    header("Content-Type: application/json; charset=" . RCUBE_CHARSET);

		echo json_encode(array(
			"username" => $rcmail->get_user_email()
		));

		exit;
	}

}
