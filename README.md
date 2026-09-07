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

Dovecot 2.4 rewrote its config syntax (named filter blocks, `%{variable}` instead of single-letter `%x` variables see [Migration documentation](https://doc.dovecot.org/main/installation/upgrade/2.3-to-2.4.html#common-modifier-syntaxes-and-their-replacements)) and old-style config is **not** backwards compatible - a leftover `%h` will silently be treated as a literal string (you'll end up with a
folder literally named `%h` on disk which doveadm writes to but apparently the plugin doesn't read from) rather than erroring, so pick the right example below for your version.

### Dovecot < 2.4

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

If you get an error `dovecot Error: Failed to set attribute: Mailbox attributes not enabled` then you need to add a line to your config (/etc/dovecot/conf.d/10-mail.conf)

```
mail_attribute_dict = file:%h/.dovecot.attributes
```

### Dovecot 2.4+

Configure the plugin

```
protocol lmtp {
  mail_plugins {
    notify = yes
    push_notification = yes
  }
}

# If notifications are also needed for LDA-based delivery, add:
protocol lda {
  mail_plugins {
    notify = yes
    push_notification = yes
  }
}

push_notification ox {
  driver = ox
  ox_url = http://dovecot:dovecot@127.0.0.1:8111/preliminary/http-notify/v1/notify
  user_from_metadata = yes
}
```

For more information checkout the [Dovecot documentation](https://doc.dovecot.org/main/core/plugins/push_notification.html)

If you get an error `dovecot Error: Failed to set attribute: Mailbox attributes not enabled` then you need to add this to your config (/etc/dovecot/conf.d/10-mail.conf)

```
mail_attribute {
  dict file {
    path = %{home}/.dovecot.attributes
  }
}
```

Note the `%{home}` - **not** the old `%h`. Using `%h` here won't error, it'll just
silently create a directory literally called `%h` under the user's home and every
notification will look like it's being skipped with "METADATA not set", which will
cost you an entire afternoon if you don't know to look for it.

### Metadata (all versions)

Configure the metadata for any users that would be using this

`doveadm mailbox metadata set -u user@example.org -s "" /private/vendor/vendor.dovecot/http-notify user=user@example.org`

You can verify it's set with `doveadm mailbox metadata get -u user@example.org -s "" /private/vendor/vendor.dovecot/http-notify`.
If that comes back empty even right after a `set`, or you find a `%h` directory sitting
under the user's home, check the `mail_attribute` config above.

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
