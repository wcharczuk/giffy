var giffyControllers = angular.module("giffyControllers", []);

giffyControllers.controller("homeController", ["$scope", "$http", 
    function($scope, $http) {
        $http.get("/api/current_user").success(function(datums) {
            $scope.current_user = datums.response;
        });
        
        $http.get("/api/images/random/5").success(function(datums) {
            $scope.images = datums.response; 
        });
        
        $scope.searchImages = function() {
            if ($scope.searchQuery) {
                $http.get("/api/search?query=" + $scope.searchQuery).success(function(datums) {
                    $scope.images = datums.response; 
                });
            } else {
                $http.get("/api/images/random/5").success(function(datums) {
                    $scope.images = datums.response; 
                });
            }
        };
    }
]);

giffyControllers.controller("addImageController", ["$scope", "$http", 
    function($scope, $http) {
        $http.get("/api/current_user").success(function(datums) {
            if (datums.response.is_logged_in) {
                $scope.current_user = datums.response;
            } else {
                window.location = "#/";    
            }
        });
    }
]);

giffyControllers.controller("imageController", ["$scope", "$http", "$routeParams", 
    function($scope, $http, $routeParams) {
        $http.get("/api/current_user").success(function(datums) {
            $scope.current_user = datums.response;
        });
        
        $scope.addTag = function() {
            $http.post("/api/tags/", {tag_value:$scope.tag_value}).success(function(datums) {
               var tag = datums.response;
               $http.post("/api/upvote/" + $scope.image.uuid + "/" + tag.uuid, {}).success(function(res) {
                    $scope.tag_value = "";
                    jQuery("#add-tag-modal").modal('hide');
                    fetchTagData();
               }); 
            });
        };
        
        jQuery('#add-tag-modal').on('shown.bs.modal', function () {
            jQuery('#tag-value').focus();
        });
        
        var fetchImageData = function() {
            delete $scope.image;

            $http.get("/api/image/" + $routeParams.image_id).success(function(datums) {
                $scope.image = datums.response;
            });
        };
        
        var fetchTagData = function() {
            delete $scope.tags;
            delete $scope.votes;
            delete $scope.voteLookup;
            
            $http.get("/api/tags/image/"+ $routeParams.image_id).success(function(datums) {
                $scope.tags = datums.response;
            });
            
            $http.get("/api/votes/image/" + $routeParams.image_id).then(function(res) {
                var voteLookup = {};
                for (var x = 0; x < res.data.response.length; x++) {
                    var vote = res.data.response[x];
                    voteLookup[vote.tag_uuid] = vote; 
                }
                $scope.votes = res.data.response;
                $scope.voteLookup = voteLookup;
            }, function(res) {});
        }
        
        $scope.deleteTag = function(tagUUID) {
            $http.delete("/api/tag/" + tagUUID).success(function(res) {
                fetchTagData(); 
            });
        }
        
        $scope.vote = function(tagUUID, isUpvote) {
            if (!$scope.hasVote(tagUUID)) {
                if (isUpvote) {
                    $http.post("/api/upvote/" + $routeParams.image_id + "/" + tagUUID, null).success(function(res) {
                        fetchTagData(); 
                    });
                } else {
                    $http.post("/api/downvote/" + $routeParams.image_id + "/" + tagUUID, null).success(function(res) {
                        fetchTagData(); 
                    });
                }
            } else {
                $http.delete("/api/vote/" + $routeParams.image_id + "/" + tagUUID).success(function() {
                   fetchTagData(); 
                });
            }
        };
        
        $scope.hasVote = function(tagUUID) {
            if ($scope.voteLookup) {
                return $scope.voteLookup[tagUUID]; //is this "truthy"??
            } else {
                return false;
            }
        };
        
        $scope.didUpvote = function(tagUUID) {
            if ($scope.voteLookup && $scope.voteLookup[tagUUID])  {
                return $scope.voteLookup[tagUUID].is_upvote;
            } else {
                return false;
            }
        };
        
        $scope.didDownvote = function(tagUUID) {
            if ($scope.voteLookup && $scope.voteLookup[tagUUID])  {
                return !$scope.voteLookup[tagUUID].is_upvote;
            } else {
                return false;
            }
        };
        
        fetchImageData();
        fetchTagData();
    }
]);

giffyControllers.controller("tagController", ["$scope", "$http", "$routeParams", 
    function($scope, $http, $routeParams) {
        $http.get("/api/current_user").success(function(datums) {
            $scope.current_user = datums.response;
        });
        
        $http.get("/api/tag/" + $routeParams.tag_id).success(function(datums) {
            $scope.tag = datums.response;
        });
        
        var fetchImageData = function() {
            $http.get("/api/images/tag/" + $routeParams.tag_id).success(function(datums) {
                $scope.images = datums.response;
            });
            
            $http.get("/api/votes/tag/" + $routeParams.tag_id).then(function(res) {
                var voteLookup = {};
                for (var x = 0; x < res.data.response.length; x++) {
                    var vote = res.data.response[x];
                    voteLookup[vote.tag_uuid] = vote; 
                }
                $scope.votes = res.data.response;
                $scope.voteLookup = voteLookup;
            }, function(res) {});
        }
        
        $scope.vote = function(tagUUID, isUpvote) {
            if (!$scope.hasVote(tagUUID)) {
                if (isUpvote) {
                    $http.post("/api/upvote/" + $routeParams.image_id + "/" + tagUUID, null).success(function(res) {
                        fetchImageData(); 
                    });
                } else {
                    $http.post("/api/downvote/" + $routeParams.image_id + "/" + tagUUID, null).success(function(res) {
                        fetchImageData();
                    });
                }
            } else {
                $http.delete("/api/vote/" + $routeParams.image_id + "/" + tagUUID).success(function() {
                   fetchImageData(); 
                });
            }
        };
        
        $scope.hasVote = function(tagUUID) {
            if ($scope.voteLookup) {
                return $scope.voteLookup[tagUUID]; //is this "truthy"??
            } else {
                return false;
            }
        };
        
        $scope.didUpvote = function(tagUUID) {
            if ($scope.voteLookup && $scope.voteLookup[tagUUID])  {
                return $scope.voteLookup[tagUUID].is_upvote;
            } else {
                return false;
            }
        };
        
        $scope.didDownvote = function(tagUUID) {
            if ($scope.voteLookup && $scope.voteLookup[tagUUID])  {
                return !$scope.voteLookup[tagUUID].is_upvote;
            } else {
                return false;
            }
        };
        
        fetchImageData();
    }
]);