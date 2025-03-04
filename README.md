# DPR - Dovecot Push Roundcube

This is a proof of concept simple push to server side event proxy, I don't imagine there's much demand for it but it solves a problem I was having.

# Theory

 * When you log into roundcube you're given a session.
 * Using reverse proxying on a url on that domain permits a SSE server to access that cookie.
 * That SSE server can run that cookie at roundcube to get your email address (and verify the cookie).
 * Dovecot put's notifications to the SSE server which forwards those notifications to anyone connected.
 * Javascript refreshes the page like normal.

# Reasoning

 * Less log noise
 * Faster notifications

# Install

YMMV but this is what I did

## Dovecot

Configure the plugin

```
protocol lmtp {
  mail_plugins = $mail_plugins notify push_notification
}

# If notifications are also needed for LDA-based delivery, add:
protocol lda {
  mail_plugins = $mail_plugins notify push_notification
}

plugin {
  push_notification_driver = ox:url=http://dovecot:dovecot@127.0.0.1:8111/preliminary/http-notify/v1/notify
}
```

For more information checkout the [Dovecot documentation](https://doc.dovecot.org/2.3/configuration_manual/push_notification/)

Configure the metadata for any users that would be using this

`doveadm mailbox metadata set -u user@example.org -s "" /private/vendor/vendor.dovecot/http-notify user=user@example.org`

If you get an error `dovecot Error: Failed to set attribute: Mailbox attributes not enabled` then you need to add a line to your config (/etc/dovecot/conf.d/10-mail.conf)

```
mail_attribute_dict = file:%h/.dovecot.attributes
```




## Roundecube

Copy sse from the roundcube-plugin directory to $ROUNDCUBE_ROOT/plugins/

Edit your config and add 'sse' to your plugins list

```
$config['plugins'] = array(
		'sse',
);
```

## Caddy

```
mail.example.org {
	root * /var/www/mail.example.org/public
	php_fastcgi 127.0.0.1:9000
	file_server
}

mail.example.org/sse/events {
	uri strip_prefix sse
	reverse_proxy localhost:8111
}
```
