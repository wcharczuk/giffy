var giffyApp = angular.module('giffyApp', [ 'ngRoute' , 'ngSanitize', 'giffyControllers' ]);


giffyApp.config(["$routeProvider", 
    function($routeProvider) {
        $routeProvider
        .when("/", {
            templateUrl: '/static/partials/home.html',
            controller: 'homeController'
        }).otherwise({ 
            templateUrl: '/static/partials/home.html',
            controller: 'homeController' 
        });
}]);