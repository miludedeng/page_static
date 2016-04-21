##页面静态化中间件
-------
可应用于动态动态服务器与代理服务器之间，经过此插件可将页面静态化，如果已存在静态页面，则直接返回。如果没有，则会生成静态页面并返回
####使用说明:
######修改conf/app.conf
`app_domain` 是动态服务器的域名或ip

`static_path` 是静态文件存放的路径

`max_expdate` 是静态文件的过期时间（默认过期时间） 

`concat_css` 可选on/off,该选项将开启将会把页面中直接引用的css直接加入页面中，以减少请求次数

*关于过期时间，不同的访问路径可以通过Header的EXPDATE设置

　　以nginx为例<br/>
　　在location中添加 proxy_set_header EXPDATE 30;<br/> 
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;*注意此处时间只能以分钟为单位

*如果页面在开发或测试时，可以在浏览器的url后拼接
`nocache=true`，该url则不会使用缓存

