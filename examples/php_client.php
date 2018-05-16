<?php
$host = getenv("GONAD_HOST");
$port = getenv("GONAD_PORT");
echo "host=$host, port=$port\n";
$s = microtime(true);
$socket = socket_create(AF_INET, SOCK_STREAM, SOL_TCP);
if ( false === $socket ) {
	die("failed to open socket");
}
$r = socket_connect($socket,$host,$port);
if ( false === $r ) {
	die("failed to connect");
}
for ( $i = 0; $i <= 1000; $i++ ) {
	$msg = "loop=$i\n";
	socket_write($socket,$msg,strlen($msg));
}
socket_close($socket);
$e = microtime(true);
echo $e - $s . "\n";

