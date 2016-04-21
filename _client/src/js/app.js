var giffyApp = angular.module('giffyApp', [ 'ngRoute', 'ngSanitize', 'ngTagsInput', 'giffy.controllers', 'giffy.directives', 'giffy.services', 'templates-dist' ]);

giffyApp.config(["$routeProvider",
	function($routeProvider) {
		$routeProvider
		.when("/", {
			templateUrl: '/static/partials/home.html',
			controller: 'homeController'
		}).when("/add_image", {
			templateUrl: '/static/partials/add_image.html',
			controller: 'addImageController'
		}).when("/search", {
			templateUrl: '/static/partials/search.html',
			controller: 'searchController'
		}).when("/search/:search_query", {
			templateUrl: '/static/partials/search.html',
			controller: 'searchController'
		}).when("/search.slack", {
			templateUrl: '/static/partials/search_slack.html',
			controller: 'searchSlackController'
		}).when("/search.slack/:search_query", {
			templateUrl: '/static/partials/search_slack.html',
			controller: 'searchSlackController'
		}).when("/images.censored", {
			templateUrl: "/static/partials/images.html",
			controller: 'imagesCensoredController'
		}).when("/add_tag", {
			templateUrl: '/static/partials/add_tag.html',
			controller: 'addTagController'
		}).when("/image/:image_id", {
			templateUrl: '/static/partials/image.html',
			controller: 'imageController'
		}).when("/tag/:tag_id", {
			templateUrl: '/static/partials/tag.html',
			controller: 'tagController'
		}).when("/user/:user_id", {
			templateUrl: '/static/partials/user.html',
			controller: 'userController'
		}).when("/users.search", {
			templateUrl: '/static/partials/users_search.html',
			controller: 'userSearchController'
		}).when("/moderation.log", {
			templateUrl: '/static/partials/moderation_log.html',
			controller: 'moderationLogController'
		}).when("/search.history", {
			templateUrl: '/static/partials/search_history.html',
			controller: 'searchHistoryController'
		}).when("/logout", {
			templateUrl: '/static/partials/logout.html',
			controller: 'logoutController'
		}).when("/about", {
			templateUrl: '/static/partials/about.html',
			controller: 'aboutController'
		}).when("/slack/complete", {
			templateUrl: '/static/partials/slack_complete.html',
			controller: 'slackCompleteController'
		}).when("/stats", {
			templateUrl: '/static/partials/stats.html',
			controller: 'statsController'
		}).otherwise({
			templateUrl: '/static/partials/home.html',
			controller: 'homeController'
		});
}]);

var copyElement = function(element) {
	var selection = window.getSelection();
	var range = document.createRange();

	range.selectNodeContents(element);
	selection.removeAllRanges();
	selection.addRange(range);

	var didSucceed = document.execCommand("copy");
	if (!didSucceed) {
		return false;
	}
	window.getSelection().removeAllRanges();
	return true;
}