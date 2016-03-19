var giffyApp = angular.module('giffyApp', [ 'ngRoute' , 'ngSanitize', 'giffy.controllers', 'giffy.directives' ]);


giffyApp.config(["$routeProvider", 
    function($routeProvider) {
        $routeProvider
        .when("/", {
            templateUrl: '/static/partials/home.html',
            controller: 'homeController'
        }).when("/add_image", {
            templateUrl: '/static/partials/add_image.html',
            controller: 'addImageController'
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
        }).when("/moderation.log", {
            templateUrl: '/static/partials/moderation_log.html',
            controller: 'moderationLogController'
        }).otherwise({ 
            templateUrl: '/static/partials/home.html',
            controller: 'homeController' 
        });
}]);