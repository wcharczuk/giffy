var giffyControllers = angular.module("giffy.controllers", []);

giffyControllers.controller("homeController", ["$scope", "$http", 
    function($scope, $http) {
        $http.get("/api/user.current").success(function(datums) {
            $scope.current_user = datums.response;
        });
        
        $http.get("/api/images/random/9").success(function(datums) {
            $scope.images = datums.response; 
            delete $scope.searchedQuery;
        });
        
        $scope.searchImages = function() {
            if ($scope.searchQuery) {
                $http.get("/api/images.search?query=" + $scope.searchQuery).success(function(datums) {
                    $scope.images = datums.response;
                    $scope.searchedQuery = $scope.searchQuery;
                });
            } else {
                $http.get("/api/images/random/9").success(function(datums) {
                    $scope.images = datums.response; 
                    delete $scope.searchedQuery;
                });
            }
        };
    }
]);

giffyControllers.controller("addImageController", ["$scope", "$http", 
    function($scope, $http) {
        $http.get("/api/user.current").success(function(datums) {
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
        $http.get("/api/user.current").success(function(datums) {
            $scope.current_user = datums.response;
        });
        
        $scope.addTag = function() {
            $http.post("/api/tags/", {tag_value:$scope.add_tag_value}).success(function(datums) {
               var tag = datums.response;
               $http.post("/api/vote.up/" + $scope.image.uuid + "/" + tag.uuid, {}).success(function(res) {
                    $scope.add_tag_value = "";
                    jQuery("#add-tag-modal").modal('hide');
                    fetchTagData();
               });
            });
        };
        
        $scope.deleteImage = function() {
            if ($scope.current_user.is_moderator) {
                $http.delete("/api/image/" + $scope.image.uuid).success(function() {
                    window.location = "/";
                });
            }
        }
        
        jQuery('#add-tag-modal').on('shown.bs.modal', function () {
            jQuery('#tag-value').focus();
        });
        
        var fetchImageData = function() {
            delete $scope.image;

            $http.get("/api/image/" + $routeParams.image_id).then(function(res) {
                $scope.image = res.data.response;
                fetchTagData();
            }, function(res) {
                window.location = "/";
            });
        };
        
        var fetchTagData = function() {
            delete $scope.tags;
            delete $scope.linkLookup;
            delete $scope.userVoteLookup;
            
            $http.get("/api/image.tags/"+ $routeParams.image_id).success(function(datums) {
                $scope.tags = datums.response;
            });
            
            $http.get("/api/image.votes/" + $routeParams.image_id).then(function(res) {
                var linkLookup = {};
                for (var x = 0; x < res.data.response.length; x++) {
                    var link = res.data.response[x];
                    linkLookup[link.tag_uuid] = link; 
                }
                $scope.linkLookup = linkLookup;
            }, function(res) {}); 
            
            if ($scope.current_user.is_logged_in) {
                $http.get("/api/user.votes.image/" + $routeParams.image_id).then(function(res) {
                    var userVoteLookup = {};
                    for (var x = 0; x < res.data.response.length; x++) {
                        var vote = res.data.response[x];
                        userVoteLookup[vote.tag_uuid] = vote; 
                    }
                    $scope.userVoteLookup = userVoteLookup;
                }, function(res) {});
            }
        }
        
        $scope.formatFileSize = function(fileSizeBytes) {
            if (fileSizeBytes > (1<<30)) {
                return (fileSizeBytes / (1<<30)).toFixed(2) + " GB";
            } 

            if (fileSizeBytes > (1 << 20)) {
                return (fileSizeBytes / (1<<20)).toFixed(2) + " MB";
            }

            if (fileSizeBytes > (1 << 10)) {
                return (fileSizeBytes / 1024).toFixed(2) + " KB";
            }

            return fileSizeBytes + " bytes"
        };

        $scope.$on("voted", function() {
           fetchTagData(); 
        });
                
        fetchImageData();
    }
]);

giffyControllers.controller("tagController", ["$scope", "$http", "$routeParams", 
    function($scope, $http, $routeParams) {
        
        // current user
        $http.get("/api/user.current").success(function(datums) {
            $scope.current_user = datums.response;
        });
        
        // tag information
        $http.get("/api/tag/" + $routeParams.tag_id).success(function(datums) {
            $scope.tag = datums.response;
            fetchVoteData();
        });
        
        var fetchVoteData = function() {
            delete $scope.linkLookup;
            delete $scope.userVoteLookup;
            
            $http.get("/api/tag.images/"+ $scope.tag.uuid).success(function(datums) {
                $scope.images = datums.response;
            });
            
            $http.get("/api/tag.votes/" + $scope.tag.uuid).then(function(res) {
                var linkLookup = {};
                for (var x = 0; x < res.data.response.length; x++) {
                    var link = res.data.response[x];
                    linkLookup[link.image_uuid] = link; 
                }
                $scope.linkLookup = linkLookup;
            }, function(res) {});
            
            if ($scope.current_user.is_logged_in) {
                $http.get("/api/user.votes.tag/" + $scope.tag.uuid).then(function(res) {
                    var userVoteLookup = {};
                    for (var x = 0; x < res.data.response.length; x++) {
                        var vote = res.data.response[x];
                        userVoteLookup[vote.image_uuid] = vote; 
                    }
                    $scope.userVoteLookup = userVoteLookup;
                }, function(res) {});
            }
        }

        $scope.$on("voted", function() {
           fetchVoteData(); 
        });
    }
]);

giffyControllers.controller("userController", ["$scope", "$http", "$routeParams", 
    function($scope, $http, $routeParams) {
        // current user
        $http.get("/api/user.current").success(function(datums) {
            $scope.current_user = datums.response;

            getUser();
        });
        
        function getUser() {
            // current user
            $http.get("/api/user/" + $routeParams.user_id).success(function(datums) {
                $scope.user = datums.response;

                $http.get("/api/user.images/" + $routeParams.user_id).success(function(datums) {
                    $scope.images = datums.response;
                });
            });
        }

        $scope.promote= function() {
            var user = $scope.user;
            user.is_moderator = !user.is_moderator;
            $http.put("/api/user/" + $routeParams.user_id, user).success(function(datums) {
                $scope.user = datums.response;
            });
        };
    }
]);

giffyControllers.controller("moderationLogController", ["$scope", "$http", "$routeParams", 
    function($scope, $http, $routeParams) {
        // current user
        $http.get("/api/user.current").success(function(datums) {
            $scope.current_user = datums.response;
        });
        
        // info
        $http.get("/api/moderation.log/pages/50/0").success(function(datums) {
            $scope.log = datums.response;
        });
    }
]);

giffyControllers.controller("userSearchController", ["$scope", "$http", 
    function($scope, $http) {
        $scope.searchUsers = function() {
            if ($scope.searchQuery) {
                $http.get("/api/users.search?query=" + $scope.searchQuery).success(function(datums) {
                    $scope.users = datums.response;
                    $scope.searchedQuery = $scope.searchQuery;
                });
            } else {
                delete $scope.users;
            }
        };
    }
]);