var giffyApp = angular.module('giffyApp', [ 'ngRoute' , 'ngSanitize', 'giffy.controllers', 'giffy.directives', 'giffy.services' ]);


giffyApp.config(["$routeProvider", 
    function($routeProvider) {
        $routeProvider
        .when("/", {
            templateUrl: '/static/partials/home.html',
            controller: 'homeController'
        }).when("/add_image", {
            templateUrl: '/static/partials/add_image.html',
            controller: 'addImageController'
        }).when("/search/:search_query", {
            templateUrl: '/static/partials/search.html',
            controller: 'searchController'
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
        }).when("/logout", {
            templateUrl: '/static/partials/logout.html',
            controller: 'logoutController'
        }).otherwise({ 
            templateUrl: '/static/partials/home.html',
            controller: 'homeController' 
        });
}]);