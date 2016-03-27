module.exports = function(grunt) {

    // Project configuration.
    grunt.initConfig({
        pkg: grunt.file.readJSON('package.json'),

        uglify: {
            app: {
                options: {
                    banner: '/*! <%= pkg.name %> <%= grunt.template.today("dd-mm-yyyy") %> */\n',
                },
                files : {
                    "_static/js/giffy.min.js" : [ "_static/js/giffy.js" ]                        
                }
            },
        },
        
        cssmin: {
            options: {
                shorthandCompacting: false,
                roundingPrecision: -1
            },
            target: {
                files: {
                    "_static/css/giffy.min.css" : ["_static/css/giffy.css"]
                }
            }
        },
        
        concat: {
            options: {
                separator: ";"
            }, 
            app: {
                src: [
                    '_bower/jquery/dist/jquery.js',
                    '_bower/bootstrap/dist/js/bootstrap.js',
                    '_bower/angular/angular.js',
                    '_bower/angular-route/angular-route.js',
                    '_static/js/app.js',
                    '_static/js/controllers.js',
                    '_static/js/directives.js',
                    '_static/js/services.js',
                ],
                dest: "_static/js/giffy.js"
            }
        },
        
        less: {            
            compile: {
                options: {
                      strictMath: true,
                      outputSourceFiles: true,
                    paths: [
                        "./",
                        "_static/less/", 
                        "_bower/bootstrap/less/",
                        "_bower/bootstrap/less/mixins"
                    ],
                },
                files: {
                    "_static/css/giffy.css" : "_static/less/giffy.less"
                }
            },
        },
        
        processhtml : {
            dist : {
                options : {
                    process: true,
                },
                files : {
                    '_static/index_compiled.html': ['_static/index.html']
                }
            }
        },
        
        cachebreaker: {
            dist: {
                options: {
                    match: ['giffy.min.js', 'giffy.min.css'],
                    position: 'append'
                },
                files: {
                    src: ['_static/index_compiled.html']
                }
            }
        }
    });

    grunt.loadNpmTasks('grunt-contrib-uglify');
    grunt.loadNpmTasks('grunt-contrib-clean');
    grunt.loadNpmTasks('grunt-contrib-copy');
    grunt.loadNpmTasks('grunt-contrib-less');
    grunt.loadNpmTasks('grunt-contrib-concat');
    grunt.loadNpmTasks('grunt-contrib-cssmin');
    grunt.loadNpmTasks('grunt-processhtml');
    grunt.loadNpmTasks('grunt-cache-breaker');
    
    grunt.registerTask(
        'build', 
        'Compiles all of the assets and copies the files to the build directory.', 
        [ 
            'concat:app',  
            'uglify:app', 
            'less:compile', 
            'cssmin',
            'processhtml:dist', 
            'cachebreaker:dist' 
        ]
    );
}