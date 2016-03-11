var giffyApp = angular.module('giffyApp', []);

giffyApp.controller('imageSearchController', ["$scope", "$http", function($scope, $http) {
    $http.get("/api/images").success(function(datums) {
        $scope.images = datums.response;
    });
}]);