var giffyApp = angular.module('giffyApp', [ 'ngRoute' , 'ngSanitize', 'giffyControllers' ]);


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
            templateUrl: '/static/partials/image_detail.html',
            controller: 'imageController'
        }).when("/tag/:tag_id", {
            templateUrl: '/static/partials/tag_detail.html',
            controller: 'tagController'
        }).otherwise({ 
            templateUrl: '/static/partials/home.html',
            controller: 'homeController' 
        });
}]);