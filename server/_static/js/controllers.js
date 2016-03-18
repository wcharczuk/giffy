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
            
            $http.get("/api/tags_for_image/"+ $routeParams.image_id).success(function(datums) {
                $scope.tags = datums.response;
            });
            
            $http.get("/api/votes/user/image/" + $routeParams.image_id).then(function(res) {
                var voteLookup = {};
                for (var x = 0; x < res.data.response.length; x++) {
                    var vote = res.data.response[x];
                    voteLookup[vote.tag_uuid] = vote; 
                }
                $scope.votes = res.data.response;
                $scope.voteLookup = voteLookup;
            }, function(res) {});
        }
        
        $scope.deleteImage = function() {
            if (confirm("Are you sure?")) {
                $http.delete("/api/image/" + $routeParams.image_id).success(function(res) {
                    window.location = "/#/";
                });
            }
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
                $http.delete("/api/votes/user/" + $routeParams.image_id + "/" + tagUUID).success(function() {
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
        // current user
        $http.get("/api/current_user").success(function(datums) {
            $scope.current_user = datums.response;
        });
        
        // tag information
        $http.get("/api/tag/" + $routeParams.tag_id).success(function(datums) {
            $scope.tag = datums.response;
            
            fetchImageData();
        });
        
        // fetch data about images that are linked to the tag
        var fetchImageData = function() {
            $http.get("/api/images_for_tag/" + $scope.tag.uuid).success(function(datums) {
               $scope.images = datums.response; 
            });
            
            $http.get("/api/links/tag/" + $scope.tag.uuid).success(function(datums) {
                var votesLookup = {};
                for (var x = 0; x < datums.response.length; x++) {
                    var summary = datums.response[x];
                    votesLookup[summary.image_uuid] = summary; 
                }
                
                $scope.votes = datums.response;
                $scope.votesLookup = votesLookup
            });
            
            // get user vote info
            $http.get("/api/votes/user/tag/" + $scope.tag.uuid).success(function(datums) {
                var userVotesLookup = {};
                for (var x = 0; x < datums.response.length; x++) {
                    var vote = datums.response[x];
                    userVotesLookup[vote.image_uuid] = vote; 
                }
                $scope.userVotes = datums.response;
                $scope.userVotesLookup = userVotesLookup;
            });
        }
        
        $scope.vote = function(imageUUID, isUpvote) {
            if (!$scope.hasVote(imageUUID)) {
                if (isUpvote) {
                    $http.post("/api/upvote/" + imageUUID + "/" + $scope.tag.uuid, null).success(function(res) {
                        fetchImageData(); 
                    });
                } else {
                    $http.post("/api/downvote/" + imageUUID + "/" + $scope.tag.uuid, null).success(function(res) {
                        fetchImageData();
                    });
                }
            } else {
                // delete a specific vote by a user for an image and a tag
                $http.delete("/api/votes/user/" + imageUUID + "/" + $scope.tag.uuid).success(function() {
                   fetchImageData(); 
                });
            }
        };
        
        $scope.hasVote = function(imageUUID) {
            if ($scope.userVotesLookup) {
                return $scope.userVotesLookup[imageUUID]; //is this "truthy"??
            } else {
                return false;
            }
        };
        
        $scope.didUpvote = function(imageUUID) {
            if ($scope.userVotesLookup && $scope.userVotesLookup[imageUUID])  {
                return $scope.userVotesLookup[imageUUID].is_upvote;
            } else {
                return false;
            }
        };
        
        $scope.didDownvote = function(imageUUID) {
            if ($scope.userVotesLookup && $scope.userVotesLookup[imageUUID])  {
                return !$scope.userVotesLookup[imageUUID].is_upvote;
            } else {
                return false;
            }
        };
        
        $scope.deleteLink = function(imageUUID) {
            //delete the link whole hog
            $http.delete("/api/link/" + imageUUID + "/" + $scope.tag.uuid).success(function(res) {
                fetchTagData(); 
            });
        }
    }
]);