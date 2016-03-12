var giffyControllers = angular.module("giffyControllers", []);

giffyControllers.controller("homeController", ["$scope", "$http", 
    function($scope, $http) {
        $http.get("/api/images").success(function(datums) {
            $scope.images = datums.response;
        });
        
        $http.get("/api/current_user").success(function(datums) {
            $scope.current_user = datums.response;
        });
        
        $scope.searchImages = function() {
            if ($scope.searchQuery && $scope.searchQuery.length > 4) {
                $http.get("/api/search?query=" + $scope.searchQuery).success(function(datums) {
                    $scope.images = datums.response; 
                });
            } else if (!$scope.searchQuery) {
                $http.get("/api/images").success(function(datums) {
                    $scope.images = datums.response;
                });
            }
        };
    }
]);
