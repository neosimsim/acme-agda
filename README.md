# acme-agda
[Agda Interaction Mode](https://agda.readthedocs.io/en/v2.6.1/tools/emacs-mode.html) for [Acme](http://acme.cat-v.org/)

You might want to add

	# existing files tagged by line number,col1-col2
	data matches '([.a-zA-Z¡-￿0-9_/\-]*[a-zA-Z¡-￿0-9_/\-]):([0-9]+),([0-9]+)-([0-9]+)'
	arg isfile     $1
	data set       $file
	attr add       addr=$2-#0+#$3-#1,$2-#0+#$4-#1
	plumb to edit
	plumb client $editor

to your plumbing rules.