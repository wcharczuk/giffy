angular.module('templates-dist', ['/static/partials/about.html', '/static/partials/add_image.html', '/static/partials/controls/footer.html', '/static/partials/controls/header.html', '/static/partials/controls/image.html', '/static/partials/controls/tag.html', '/static/partials/controls/username.html', '/static/partials/controls/vote_button.html', '/static/partials/home.html', '/static/partials/image.html', '/static/partials/logout.html', '/static/partials/moderation_log.html', '/static/partials/search.html', '/static/partials/slack_complete.html', '/static/partials/stats.html', '/static/partials/tag.html', '/static/partials/user.html', '/static/partials/users_search.html']);

angular.module("/static/partials/about.html", []).run(["$templateCache", function($templateCache) {
  $templateCache.put("/static/partials/about.html",
    "<giffy-header/>\n" +
    "<div class=\"row page-body\">\n" +
    "	<div class=\"col-md-6 col-md-offset-3 align-center\">\n" +
    "		<div>\n" +
    "			<img ng-src=\"{{image.s3_read_url}}\"/>\n" +
    "		</div>\n" +
    "		<div class=\"page-header\">\n" +
    "			<h2>What is `giffy`?</h2>\n" +
    "		</div>\n" +
    "		<div>Giffy is a gif search service where you, the user, get to vote on which tags apply to what images.</div>\n" +
    "		<div class=\"page-header\">\n" +
    "			<h2>Why make this?</h2>\n" +
    "		</div>\n" +
    "		<div>Because other gif search services left a lot to be desired.</div>\n" +
    "		<div class=\"page-header\">\n" +
    "			<h2>Why can't I add images?</h2>\n" +
    "		</div>\n" +
    "		<div>Because you're not a mod. If you want to be a moderator, <a href=\"https://twitter.com/willcharczuk\">ping me</a>.</div>\n" +
    "		<div class=\"page-header\">\n" +
    "			<h2>Help! The results suck!</h2>\n" +
    "		</div>\n" +
    "		<div>Then you failed, vote on tags or add new ones and make it better.</div>\n" +
    "	</div>\n" +
    "</div>\n" +
    "<giffy-footer/>");
}]);

angular.module("/static/partials/add_image.html", []).run(["$templateCache", function($templateCache) {
  $templateCache.put("/static/partials/add_image.html",
    "<giffy-header/>\n" +
    "<div class=\"row page-header\">\n" +
    "	<div class=\"page-header align-center\">\n" +
    "		<h3>Add an image</h3>\n" +
    "	</div>\n" +
    "	<div class=\"col-xs-10 col-xs-offset-1 col-sm-10 col-sm-offset-1 col-md-10 col-md-offset-1\">\n" +
    "		<iframe id=\"image-upload-frame\" src=\"/images/upload\" frameBorder=\"0\" seamless=\"seamless\" scrolling=\"no\" />\n" +
    "	</div>\n" +
    "</div>\n" +
    "<giffy-footer/>");
}]);

angular.module("/static/partials/controls/footer.html", []).run(["$templateCache", function($templateCache) {
  $templateCache.put("/static/partials/controls/footer.html",
    "<footer id=\"giffy-footer\" class=\"giffy-footer\">\n" +
    "	<div class=\"row\">\n" +
    "		<div class=\"col-md-10 col-md-offset-1 align-center\">\n" +
    "			<div class=\"footer-section\">\n" +
    "				Giffy was created by <a href=\"http://twitter.com/willcharczuk\">@willcharczuk</a>\n" +
    "			</div>\n" +
    "			<div class=\"footer-section\">|</div>\n" +
    "			<div class=\"footer-section\">\n" +
    "				<a href=\"/#/\">Home</a>\n" +
    "			</div>\n" +
    "			<div class=\"footer-section\">\n" +
    "				<a href=\"/#/about\">About</a>\n" +
    "			</div>\n" +
    "			<div class=\"footer-section\" ng-if=\"currentUser.is_admin\">\n" +
    "				<a href=\"/#/users.search\">Users</a>\n" +
    "			</div>\n" +
    "			<div class=\"footer-section\">\n" +
    "				<a href=\"/#/moderation.log\">Moderation Log</a>\n" +
    "			</div>\n" +
    "			<div class=\"footer-section\">\n" +
    "				<a href=\"/#/stats\">Stats</a>\n" +
    "			</div>\n" +
    "			<div class=\"footer-section\">\n" +
    "				<a href=\"https://slack.com/oauth/authorize?scope=commands&client_id=12971878642.28451191280\"><img alt=\"Add to Slack\" height=\"40\" width=\"139\" src=\"https://platform.slack-edge.com/img/add_to_slack.png\" srcset=\"https://platform.slack-edge.com/img/add_to_slack.png 1x, https://platform.slack-edge.com/img/add_to_slack@2x.png 2x\" /></a>\n" +
    "			</div>\n" +
    "		</div>\n" +
    "	</div>\n" +
    "</footer>");
}]);

angular.module("/static/partials/controls/header.html", []).run(["$templateCache", function($templateCache) {
  $templateCache.put("/static/partials/controls/header.html",
    "<div id=\"giffy-header\" class=\"row\">\n" +
    "	<div class=\"col-sm-12 col-md-10 col-md-offset-1\">\n" +
    "		<div class=\"row\">\n" +
    "			<div id=\"giffy-header-search-cell\" class=\"col-xs-12 col-sm-5 col-md-5\">\n" +
    "				<a id=\"giffy-search\" href=\"/#/search\">\n" +
    "					<span class=\"glyphicon glyphicon-search\"></span> Search\n" +
    "				</a>\n" +
    "			</div>\n" +
    "			<div id=\"giffy-header-logo-cell\" class=\"col-xs-12 col-sm-2 col-md-2\">\n" +
    "				<div id=\"giffy-header-logo\">\n" +
    "					<a href=\"/#/\"><img id=\"giffy-logo\" src=\"/static/images/logo.png\"/></a>\n" +
    "				</div>\n" +
    "			</div>\n" +
    "			<div id=\"giffy-header-add-image-cell\" class=\"col-xs-12 col-sm-5 col-md-5\">\n" +
    "				<a id=\"#giffy-header-add-image\" class=\"btn btn-primary\" ng-if=\"currentUser.is_moderator\" href=\"/#/add_image\" >\n" +
    "					<span class=\"glyphicon glyphicon-upload\" aria-hidden=\"true\"></span>Add Image\n" +
    "				</a>\n" +
    "			</div>\n" +
    "		</div>\n" +
    "	</div>\n" +
    "</div>\n" +
    "<div id=\"giffy-sub-header\" class=\"row\">\n" +
    "	<div class=\"col-sm-12 col-md-10 col-sm-offset-1 col-md-offset-1 align-right\">\n" +
    "		<user-detail user=\"currentUser\" ng-if=\"currentUser.is_logged_in\"></user-detail>\n" +
    "		<a id=\"header-login\" class=\"btn btn-danger btn-sm\" href=\"{{currentUser.google_login_url}}\" ng-if=\"!currentUser.is_logged_in\">Login with Google</a>\n" +
    "		<a id=\"header-logout\" class=\"btn btn-default btn-sm\" ng-if=\"currentUser.is_logged_in\" href=\"/#/logout\">Log Out</a>\n" +
    "	</div>\n" +
    "</div>");
}]);

angular.module("/static/partials/controls/image.html", []).run(["$templateCache", function($templateCache) {
  $templateCache.put("/static/partials/controls/image.html",
    "<div class=\"image-element\">\n" +
    "	<img class=\"image-element-img\" ng-src=\"{{image.s3_read_url}}\"/>\n" +
    "	<a class=\"image-element-link\" href=\"/#/image/{{image.uuid}}\"></a>\n" +
    "	<div class=\"image-tags\">\n" +
    "		<a class=\"label label-primary\" href=\"/#/tag/{{tag.tag_value}}\" ng-repeat=\"tag in image.tags\">#{{tag.tag_value}}</a>\n" +
    "	</div>\n" +
    "</a>");
}]);

angular.module("/static/partials/controls/tag.html", []).run(["$templateCache", function($templateCache) {
  $templateCache.put("/static/partials/controls/tag.html",
    "<div class=\"tag\">\n" +
    "	<div class=\"tag-part tag-detail-link btn btn-info btn-xs\"><a href=\"#/tag/{{tag.uuid}}\">{{tag.tag_value}}</a>({{tag.votes_total}})</div>\n" +
    "	<div class=\"tag-part tag-vote tag-upvote\" ng-if=\"currentUser.is_logged_in\" data-tag-id=\"{{tag.uuid}}\" data-is-upvote=\"true\">up</div>\n" +
    "	<div class=\"tag-part tag-vote tag-downvote\" ng-if=\"currentUser.is_logged_in\" data-tag-id=\"{{tag.uuid}}\" data-is-upvote=\"true\">down</div>\n" +
    "</div>");
}]);

angular.module("/static/partials/controls/username.html", []).run(["$templateCache", function($templateCache) {
  $templateCache.put("/static/partials/controls/username.html",
    "<span class=\"user-detail\">\n" +
    "	<a href=\"/#/user/{{user.uuid}}\">{{user.username}}</a>\n" +
    "	<div class=\"user-detail-badges\">\n" +
    "		<span class=\"label label-info\" ng-if=\"user.is_admin\">Admin</span>\n" +
    "		<span class=\"label label-primary\" ng-if=\"user.is_moderator\">Moderator</span>\n" +
    "		<span class=\"label label-danger\" ng-if=\"user.is_banned\">Banned</span>\n" +
    "	</div>\n" +
    "</span>");
}]);

angular.module("/static/partials/controls/vote_button.html", []).run(["$templateCache", function($templateCache) {
  $templateCache.put("/static/partials/controls/vote_button.html",
    "<div class=\"vote-button\"><div class=\"vote-button-part vote-count\" ng-class=\"{'singleton': isOnlyVoteCount()}\">{{link.votes_total}}</div><div class=\"vote-button-part vote-controls\" ng-if=\"userIsLoggedIn()\"><div class=\"vote-arrow vote-upvote\" ng-class=\"{voted: didUpvote()}\" ng-click=\"vote(true)\">+</div><div class=\"vote-arrow vote-downvote\" ng-class=\"{voted: didDownvote()}\" ng-click=\"vote(false)\">-</div></div><div class=\"vote-button-part vote-detail-link\" ng-if=\"type === 'tag'\"><a href=\"{{detailURL()}}\">{{detailValue()}}</a></div><div class=\"vote-button-part vote-remove\" ng-if=\"canEdit()\" ng-click=\"delete()\">x</div></div>");
}]);

angular.module("/static/partials/home.html", []).run(["$templateCache", function($templateCache) {
  $templateCache.put("/static/partials/home.html",
    "<giffy-header/>\n" +
    "\n" +
    "<div id=\"giffy-home\" class=\"row page-body\">\n" +
    "	<div class=\"col-xs-10 col-xs-offset-1 col-sm-10 col-sm-offset-1 col-md-10 col-md-offset-1\">\n" +
    "		<div class=\"row\">\n" +
    "			<div class=\"col-md-12\">\n" +
    "				<form>\n" +
    "					<input type=\"text\" tabindex=\"0\" class=\"giffy-search-box form-control input-lg\" placeholder=\"Search\" ng-model=\"searchQuery\" ng-enter=\"searchImages(searchQuery)\" required/>\n" +
    "				</form>\n" +
    "			</div>\n" +
    "		</div>\n" +
    "\n" +
    "		<div class=\"row\">\n" +
    "			<div class=\"col-md-12\">\n" +
    "				<div class=\"page-header align-center\">\n" +
    "					<h3>Here are some random images</h3>\n" +
    "				</div>\n" +
    "				<giffy-image class=\"image-result\" image=\"image\"  ng-repeat=\"image in images\"></giffy-image>\n" +
    "			</div>\n" +
    "		</div>\n" +
    "	</div>\n" +
    "</div>\n" +
    "\n" +
    "<giffy-footer/>");
}]);

angular.module("/static/partials/image.html", []).run(["$templateCache", function($templateCache) {
  $templateCache.put("/static/partials/image.html",
    "<giffy-header/>\n" +
    "\n" +
    "<div id=\"giffy-image\">\n" +
    "	<div class=\"row image\">\n" +
    "		<div class=\"col-md-4 col-md-offset-4 align-center\">\n" +
    "			<img class=\"giffy-image-detail\" ng-src=\"{{image.s3_read_url}}\" alt=\"{{image.display_name}}\"/>\n" +
    "		</div>\n" +
    "	</div>\n" +
    "\n" +
    "	<div class=\"row\">\n" +
    "		<div class=\"col-xs-10 col-xs-offset-1 col-sm-10 col-sm-offset-1 col-md-10 col-md-offset-1 align-center\">\n" +
    "			<div class=\"page-header\">\n" +
    "				<h1>Tags</h1>\n" +
    "			</div>\n" +
    "			<ul class=\"image-tags\">\n" +
    "				<li ng-repeat=\"tag in tags\">\n" +
    "					<vote-button type=\"'tag'\" current-user=\"currentUser\" link=\"linkLookup[tag.uuid]\" user-vote=\"userVoteLookup[tag.uuid]\" object=\"tag\"></vote-button>\n" +
    "				</li>\n" +
    "				<li ng-if=\"currentUser.is_logged_in && !currentUser.is_banned\">\n" +
    "					<button class=\"btn btn-default btn-lg\" data-toggle=\"modal\" data-target=\"#add-tag-modal\">+</button>\n" +
    "				</li>\n" +
    "			</ul>\n" +
    "		</div>\n" +
    "	</div>\n" +
    "\n" +
    "	<div class=\"row image-details\">\n" +
    "		<div class=\"col-xs-10 col-xs-offset-1 col-sm-10 col-sm-offset-1 col-md-10 col-md-offset-1\">\n" +
    "			<div class=\"row\" ng-if=\"currentUser.is_moderator\">\n" +
    "				<div class=\"col-sm-6 col-md-6 col-lg-5 align-right\">Status:</div>\n" +
    "				<div class=\"col-sm-6 col-md-6 col-lg-7 align-left\">\n" +
    "					<span class=\"label label-danger\" ng-if=\"image.is_censored\">Censored</span>\n" +
    "					<span class=\"label label-info\" ng-if=\"!image.is_censored\">Live</span>\n" +
    "				</div>\n" +
    "			</div>\n" +
    "			<div class=\"row\">\n" +
    "				<div class=\"col-sm-6 col-md-6 col-lg-5 align-right\">Slack Command:</div>\n" +
    "				<div class=\"col-sm-6 col-md-6 col-lg-7 align-left\"><a id=\"slack-command-link\" href=\"#\" title=\"Copy Code To Clipboard\"><code>{{slackCommand}}</code></a></div>\n" +
    "			</div>\n" +
    "			<div class=\"row\">\n" +
    "				<div class=\"col-sm-6 col-md-6 col-lg-5 align-right\">Created:</div>\n" +
    "				<div class=\"col-sm-6 col-md-6 col-lg-7 align-left\">{{image.created_utc | date:short}}</div>\n" +
    "			</div>\n" +
    "			<div class=\"row\">\n" +
    "				<div class=\"col-sm-6 col-md-6 col-lg-5 align-right\">Dimensions:</div>\n" +
    "				<div class=\"col-sm-6 col-md-6 col-lg-7 align-left\">{{image.width}}x{{image.height}} px</div>\n" +
    "			</div>\n" +
    "			<div class=\"row\">\n" +
    "				<div class=\"col-sm-6 col-md-6 col-lg-5 align-right\">File Size:</div>\n" +
    "				<div class=\"col-sm-6 col-md-6 col-lg-7 align-left\">{{formatFileSize(image.file_size)}}</div>\n" +
    "			</div>\n" +
    "			<div class=\"row\">\n" +
    "				<div class=\"col-sm-6 col-md-6 col-lg-5 align-right\">Added By:</div>\n" +
    "				<div class=\"col-sm-6 col-md-6 col-lg-7 align-left\"><user-detail user=\"image.created_by\"></user-detail></div>\n" +
    "			</div>\n" +
    "			<div class=\"row\" ng-if=\"currentUser.is_moderator || currentUser.user_uuid == image.created_by.uuid\">\n" +
    "				<div class=\"col-sm-6 col-md-6 col-lg-5 align-right\">Moderator Controls:</div>\n" +
    "				<div class=\"col-sm-6 col-md-6 col-lg-7 align-left\">\n" +
    "					<button class=\"btn btn-default btn-xs\" ng-click=\"censorImage()\" ng-if=\"!image.is_censored\">Censor Image</button>\n" +
    "					<button class=\"btn btn-default btn-xs\" ng-click=\"uncensorImage()\" ng-if=\"image.is_censored\">Uncensor Image</button>\n" +
    "					<button class=\"btn btn-danger btn-xs\" ng-click=\"deleteImage()\">Delete Image</button>\n" +
    "				</div>\n" +
    "			</div>\n" +
    "		</div>\n" +
    "	</div>\n" +
    "\n" +
    "	<div id=\"add-tag-modal\" class=\"modal\" tabindex=\"-1\" role=\"dialog\">\n" +
    "		<div class=\"modal-dialog\">\n" +
    "			<div class=\"modal-content\">\n" +
    "				<div class=\"modal-header\">\n" +
    "					<button type=\"button\" class=\"close\" data-dismiss=\"modal\" aria-label=\"Close\"><span aria-hidden=\"true\">&times;</span></button>\n" +
    "					<h4 class=\"modal-title\">Add Tag(s)</h4>\n" +
    "				</div>\n" +
    "				<div class=\"modal-body\">\n" +
    "					<div id=\"add-tag\" class=\"add-tag\">\n" +
    "						<!-- put validation here etc. -->\n" +
    "						<form ng-submit=\"addTags()\" ng-enter=\"addTags()\">\n" +
    "							<div class=\"form-group row\">\n" +
    "								<div class=\"col-md-10 col-md-offset-1\">\n" +
    "									<tags-input id=\"tagsInput\" tabindex=\"0\" ng-model=\"newTags\" spellcheck=\"false\" replace-spaces-with-dashes=\"false\" placeholder=\"lolz\" on-tag-added=\"tagAddedHandler()\">\n" +
    "										<auto-complete source=\"searchTags($query)\" select-first-match=\"false\"></auto-complete>\n" +
    "									</tags-input>\n" +
    "								</div>\n" +
    "							</div>\n" +
    "						</form>\n" +
    "					</div>\n" +
    "				</div>\n" +
    "				<div class=\"modal-footer\">\n" +
    "					<button type=\"button\" class=\"btn btn-default\" data-dismiss=\"modal\" tabindex=\"11\">Close</button>\n" +
    "					<button type=\"button\" class=\"btn btn-primary\" ng-click=\"addTags()\" tabindex=\"10\">Add Tag(s)</button>\n" +
    "				</div>\n" +
    "			</div>\n" +
    "		</div>\n" +
    "	</div>\n" +
    "</div>\n" +
    "\n" +
    "<giffy-footer/>");
}]);

angular.module("/static/partials/logout.html", []).run(["$templateCache", function($templateCache) {
  $templateCache.put("/static/partials/logout.html",
    "<giffy-header/>\n" +
    "<div class=\"row\">\n" +
    "	<div class=\"col-md-10 col-md-offset-1\">\n" +
    "		<h4>Logging Out...</h4>\n" +
    "	</div>\n" +
    "</div>\n" +
    "<giffy-footer/>");
}]);

angular.module("/static/partials/moderation_log.html", []).run(["$templateCache", function($templateCache) {
  $templateCache.put("/static/partials/moderation_log.html",
    "<giffy-header/>\n" +
    "\n" +
    "<div class=\"row\">\n" +
    "	<div class=\"col-xs-10 col-xs-offset-1 col-sm-10 col-sm-offset-1 col-md-10 col-md-offset-1\">\n" +
    "		<div class=\"page-header\">\n" +
    "			<h1>Moderation Log</h1>\n" +
    "		</div>\n" +
    "		<table id=\"moderation-log\" class=\"table table-responsive\">\n" +
    "			<thead>\n" +
    "				<tr>\n" +
    "					<th>User</th><th>Time</th><th>Verb</th><th>Object</th><th>Noun</th><th>Secondary Noun</th>\n" +
    "				</tr>\n" +
    "			</thead>\n" +
    "			<tbody>\n" +
    "				<tr ng-repeat=\"entry in log\">\n" +
    "					<td><user-detail user=\"entry.moderator\"/></td>\n" +
    "					<td>{{entry.timestamp_utc | date:'short' }}</td>\n" +
    "					<td>{{entry.verb}}</td>\n" +
    "					<td>{{entry.object}}</td>\n" +
    "					<td>\n" +
    "						<a href=\"/#/user/{{entry.user.uuid}}\" ng-if=\"entry.object === 'user'\">{{entry.user.username}}</a>\n" +
    "						<a href=\"/#/image/{{entry.image.uuid}}\" ng-if=\"entry.object === 'image' || entry.object === 'link'\">{{entry.image.display_name}}</a>\n" +
    "						<a href=\"/#/tag/{{entry.tag.tag_value}}\" ng-if=\"entry.object === 'tag'\">{{entry.tag.tag_value}}</a>\n" +
    "					</td>\n" +
    "					<td>\n" +
    "						<a href=\"/#/tag/{{entry.tag.tag_value}}\" ng-if=\"entry.object === 'link'\">{{entry.tag.tag_value}}</a>\n" +
    "					</td>\n" +
    "				</tr> \n" +
    "			</tbody>\n" +
    "		</table>\n" +
    "		<div class=\"row\">\n" +
    "			<div class=\"col-md-6 col-md-offset-3 align-center\">\n" +
    "				<ul class=\"pager\">\n" +
    "					<li><a ng-class=\"{disabled: !hasPreviousPage()}\" ng-click=\"previousPage()\">Previous</a></li>\n" +
    "					<li><a ng-class=\"{disabled: !hasNextPage()}\" ng-click=\"nextPage()\">Next</a></li>\n" +
    "				</ul>\n" +
    "			</div>\n" +
    "		</div>\n" +
    "	</div>\n" +
    "</div>\n" +
    "\n" +
    "<giffy-footer/>");
}]);

angular.module("/static/partials/search.html", []).run(["$templateCache", function($templateCache) {
  $templateCache.put("/static/partials/search.html",
    "<giffy-header/>\n" +
    "\n" +
    "<div id=\"giffy-search\" class=\"row page-body\">\n" +
    "	<div class=\"col-md-10 col-md-offset-1\">\n" +
    "		<div class=\"row\">\n" +
    "			<div class=\"col-md-12\">\n" +
    "				<form>\n" +
    "					<input type=\"text\" tabindex=\"0\" class=\"giffy-search-box form-control input-lg\" placeholder=\"Search\" ng-model=\"searchQuery\" ng-enter=\"searchImages(searchQuery)\" required/>\n" +
    "				</form>\n" +
    "			</div>\n" +
    "		</div>\n" +
    "\n" +
    "		<div class=\"row\">\n" +
    "			<div class=\"col-md-12\">\n" +
    "				<div class=\"page-header align-center\">\n" +
    "					<h3 ng-if=\"!!searchedQuery\">Image Search Results For: <span class=\"giffy-search-query\">{{searchedQuery}}</h3>\n" +
    "				</div>\n" +
    "				<giffy-image class=\"image-result\" image=\"image\"  ng-repeat=\"image in images\"></giffy-image>\n" +
    "			</div>\n" +
    "		</div>\n" +
    "	</div>\n" +
    "</div>\n" +
    "\n" +
    "<giffy-footer/>");
}]);

angular.module("/static/partials/slack_complete.html", []).run(["$templateCache", function($templateCache) {
  $templateCache.put("/static/partials/slack_complete.html",
    "<giffy-header/>\n" +
    "<div class=\"row page-body\">\n" +
    "	<div class=\"col-md-6 col-md-offset-3 align-center\">\n" +
    "		<img src=\"https://assets.brandfolder.com/ubhnmsn4/view.png\"/>\n" +
    "		<div class=\"page-header align-center\">\n" +
    "			<h2>Giffy has been added to slack!</h2>\n" +
    "		</div>\n" +
    "	</div>\n" +
    "</div>\n" +
    "<giffy-footer/>");
}]);

angular.module("/static/partials/stats.html", []).run(["$templateCache", function($templateCache) {
  $templateCache.put("/static/partials/stats.html",
    "<giffy-header/>\n" +
    "<div id=\"giffy-stats\" class=\"row page-body\">\n" +
    "	<div class=\"col-md-10 col-md-offset-1\">\n" +
    "		<div class=\"row\">\n" +
    "			<div class=\"col-md-4\">\n" +
    "				<div class=\"page-header\">\n" +
    "					<h4>Users</h4>\n" +
    "				</div>\n" +
    "				<div class=\"giffy-stat\">\n" +
    "					{{stats.user_count}}\n" +
    "				</div>\n" +
    "			</div>\n" +
    "			<div class=\"col-md-4\">\n" +
    "				<div class=\"page-header\">\n" +
    "					<h4>Images</h4>\n" +
    "				</div>\n" +
    "				<div class=\"giffy-stat\">\n" +
    "					{{stats.image_count}}\n" +
    "				</div>\n" +
    "			</div>\n" +
    "			<div class=\"col-md-4\">\n" +
    "				<div class=\"page-header\">\n" +
    "					<h4>Tags</h4>\n" +
    "				</div>\n" +
    "				<div class=\"giffy-stat\">\n" +
    "					{{stats.tag_count}}\n" +
    "				</div>\n" +
    "			</div>\n" +
    "		</div>\n" +
    "		<div class=\"row\">\n" +
    "			<div class=\"col-md-4\">\n" +
    "				<div class=\"page-header\">\n" +
    "					<h4>Total Karma</h4>\n" +
    "				</div>\n" +
    "				<div class=\"giffy-stat\">\n" +
    "					{{stats.karma_total}}\n" +
    "				</div>\n" +
    "			</div>\n" +
    "			\n" +
    "			<div class=\"col-md-4\">\n" +
    "				<div class=\"page-header\">\n" +
    "					<h4>Orphaned Tags</h4>\n" +
    "				</div>\n" +
    "				<div class=\"giffy-stat\">\n" +
    "					{{stats.orphaned_tag_count}}\n" +
    "				</div>\n" +
    "			</div>\n" +
    "		</div>\n" +
    "	</div>\n" +
    "</div>\n" +
    "<giffy-footer/>");
}]);

angular.module("/static/partials/tag.html", []).run(["$templateCache", function($templateCache) {
  $templateCache.put("/static/partials/tag.html",
    "<giffy-header/>\n" +
    "\n" +
    "<div class=\"row page-body\">\n" +
    "	<div class=\"col-md-10 col-md-offset-1 align-center\">\n" +
    "		<div class=\"page-header\">\n" +
    "			<h1>{{tag.tag_value}}</h1>\n" +
    "		</div>\n" +
    "		<div id=\"tag-image-results\">\n" +
    "			<div class=\"tag-image-result\" ng-repeat=\"image in images\">\n" +
    "				<giffy-image image=\"image\"></giffy-image>\n" +
    "				<vote-button type=\"'image'\" current-user=\"currentUser\" link=\"linkLookup[image.uuid]\" user-vote=\"userVoteLookup[image.uuid]\" object=\"image\"></vote-button>\n" +
    "			</div>\n" +
    "		</div>\n" +
    "	</div>\n" +
    "</div>\n" +
    "\n" +
    "<giffy-footer/>");
}]);

angular.module("/static/partials/user.html", []).run(["$templateCache", function($templateCache) {
  $templateCache.put("/static/partials/user.html",
    "<giffy-header/>\n" +
    "\n" +
    "<div id=\"giffy-user\" class=\"row page-body\">\n" +
    "	<div class=\"col-xs-10 col-xs-offset-1 col-sm-10 col-sm-offset-1 col-md-8 col-md-offset-2\">\n" +
    "		<div class=\"row\">\n" +
    "			<div class=\"col-md-12\">\n" +
    "				<div class=\"page-header\">\n" +
    "					<h1>{{user.username}}</h1>\n" +
    "				</div>\n" +
    "				<div class=\"info-row row\">\n" +
    "					<div class=\"col-md-2\">Username:</div>\n" +
    "					<div class=\"col-md-4\">{{user.username}}</div>\n" +
    "					<div class=\"col-md-2\">Roles:</div>\n" +
    "					<div class=\"col-md-4\">\n" +
    "						<span class=\"label label-info ng-scope\" ng-if=\"user.is_admin\">Admin</span>\n" +
    "						<span class=\"label label-primary ng-scope\" ng-if=\"user.is_moderator\">Moderator</span>\n" +
    "						<span class=\"label label-danger ng-scope\" ng-if=\"user.is_banned\">Banned</span>\n" +
    "					</div>\n" +
    "				</div>\n" +
    "				<div class=\"info-row row\">\n" +
    "					<div class=\"col-md-2\">Name:</div>\n" +
    "					<div class=\"col-md-4\">{{user.first_name}} {{user.last_name}}</div>\n" +
    "\n" +
    "					<div class=\"col-md-2\">Joined:</div>\n" +
    "					<div class=\"col-md-4\">{{user.created_utc}}</div>\n" +
    "				</div>\n" +
    "				<div class=\"info-row row\" ng-if=\"currentUser.is_admin\">\n" +
    "					<div class=\"col-md-2\">Admin Controls:</div>\n" +
    "					<div class=\"col-md-10\">\n" +
    "						<button class=\"btn btn-primary btn-xs\" ng-click=\"promote()\" ng-disabled=\"user.is_admin\">Moderator</button>\n" +
    "						<button class=\"btn btn-danger btn-xs\" ng-click=\"ban()\" ng-disabled=\"user.is_admin\">Ban</button>\n" +
    "					</div>\n" +
    "				</div>\n" +
    "			</div>\n" +
    "		</div>\n" +
    "		<div class=\"row\">\n" +
    "			<div class=\"col-md-12 image-search-results\">\n" +
    "				<div class=\"page-header\">\n" +
    "					<h4>Images they've added:</h4>\n" +
    "				</div>\n" +
    "				<giffy-image class=\"image-result\" image=\"image\" ng-repeat=\"image in images\"></giffy-image>\n" +
    "			</div>\n" +
    "		</div>\n" +
    "	</div>\n" +
    "</div>\n" +
    "\n" +
    "<giffy-footer/>");
}]);

angular.module("/static/partials/users_search.html", []).run(["$templateCache", function($templateCache) {
  $templateCache.put("/static/partials/users_search.html",
    "<giffy-header/>\n" +
    "\n" +
    "<div class=\"row page-body\">\n" +
    "	<div class=\"col-xs-10 col-xs-offset-1 col-sm-10 col-sm-offset-1 col-md-10 col-md-offset-1\">\n" +
    "		<div class=\"page-header\">\n" +
    "			<h1>Search Users</h1>\n" +
    "		</div>\n" +
    "		<form>\n" +
    "			<input id=\"giffy-user-search-bar\" type=\"text\" tabindex=\"0\" class=\"form-control input-lg\" placeholder=\"Search\" ng-model=\"searchQuery\" ng-enter=\"searchUsers()\" required/>\n" +
    "		</form>\n" +
    "		<table class=\"table table-responsive\" style=\"margin-top:25px;\">\n" +
    "			<thead ng-if=\"!!users && users.length > 0\">\n" +
    "				<tr>\n" +
    "					<th>Username</th>\n" +
    "					<th>Created</th>\n" +
    "					<th>First Name</th>\n" +
    "					<th>Last Name</th>\n" +
    "					<th>Roles / Status</th>\n" +
    "				</tr>\n" +
    "			</thead>\n" +
    "			<tbody>\n" +
    "				<tr ng-repeat=\"user in users\">\n" +
    "					<td><a href=\"/#/user/{{user.uuid}}\">{{user.username}}</a></td>\n" +
    "					<td>{{user.created_utc | date:short}}</td>\n" +
    "					<td>{{user.first_name}}</td>\n" +
    "					<td>{{user.last_name}}</td>\n" +
    "					<td>\n" +
    "						<span class=\"label label-info ng-scope\" ng-if=\"user.is_admin\">Admin</span>\n" +
    "						<span class=\"label label-primary ng-scope\" ng-if=\"user.is_moderator\">Moderator</span>\n" +
    "						<span class=\"label label-danger ng-scope\" ng-if=\"user.is_banned\">Banned</span>\n" +
    "					</td>\n" +
    "				</tr>\n" +
    "			</tbody>\n" +
    "		</table>\n" +
    "	</div>\n" +
    "</div>\n" +
    "\n" +
    "<giffy-footer/>");
}]);
