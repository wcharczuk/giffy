var giffyControllers = angular.module("giffyControllers", []);

giffyControllers.controller("homeController", ["$scope", "$http", 
    function($scope, $http) {
        $http.get("/api/current_user").success(function(datums) {
            $scope.current_user = datums.response;
        });
        
        $scope.searchImages = function() {
            $http.get("/api/search?query=" + $scope.searchQuery).success(function(datums) {
                $scope.images = datums.response; 
            });
        };
    }
]);

giffyControllers.controller("addImageController", ["$scope", "$http", 
    function($scope, $http) {
        $http.get("/api/current_user").success(function(datums) {
            $scope.current_user = datums.response;
        }).error(function() {
            window.location = "#/";
        });
    }
]);

giffyControllers.controller("imageController", ["$scope", "$http", "$routeParams", 
    function($scope, $http, $routeParams) {
        $http.get("/api/current_user").success(function(datums) {
            $scope.current_user = datums.response;
        });
        
        $http.get("/api/image/" + $routeParams.image_id).success(function(datums) {
            $scope.image = datums.response;
        });
        
        $http.get("/api/images/tags/"+ $routeParams.image_id).success(function(datums) {
            $scope.tags = datums.response; 
        });
    }
]);