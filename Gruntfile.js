module.exports = function(grunt) {

	// Project configuration.
	grunt.initConfig({
		pkg: grunt.file.readJSON('package.json'),

		uglify: {
			app: {
				options: {
					banner: '/*! <%= pkg.name %> <%= grunt.template.today("yyyy-mm-dd") %> */\n',
				},
				files : {
					"_client/dist/js/giffy.min.js" : [ "_client/dist/js/giffy.js" ]
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
					"_client/dist/css/giffy.min.css" : ["_client/dist/css/giffy.css"]
				}
			}
		},

		concat: {
			options: {
				separator: ";"
			},
			app: {
				src: [
					'_client/bower/jquery/dist/jquery.js',
					'_client/bower/bootstrap/dist/js/bootstrap.js',
					'_client/bower/angular/angular.js',
					'_client/bower/angular-route/angular-route.js',
					'_client/bower/angular-sanitize/angular-sanitize.js',
					'_client/bower/ng-tags-input/ng-tags-input.js',
					'_client/src/js/app.js',
					'_client/src/js/controllers.js',
					'_client/src/js/directives.js',
					'_client/src/js/services.js',
					'_client/dist/js/partials.js',

				],
				dest: "_client/dist/js/giffy.js"
			}
		},

		html2js: {
			options: {
				base: '_client/src/',
				rename: function(name) {
					return '/static/' + name;
				}
			},
			dist: {
				src: ["_client/src/partials/**/*.html"],
				dest: "_client/dist/js/partials.js"
			}
		},

		less: {
			compile: {
				options: {
					  strictMath: true,
					  outputSourceFiles: true,
					paths: [
						"_client/",
						"_client/src/less/",
						"_client/bower/bootstrap/less/",
						"_client/bower/bootstrap/less/mixins"
					],
				},
				files: {
					"_client/dist/css/giffy.css" : "_client/src/less/giffy.less"
				}
			},
		},

		copy: {
		  dist: {
		   files: [
			   { src: "_client/bower/jquery/dist/jquery.min.js", dest: "_client/dist/js/jquery.min.js" },
			   { src: "_client/bower/bootstrap/dist/js/bootstrap.min.js", dest: "_client/dist/js/bootstrap.min.js" },
			   { expand: true, flatten: true, src: "_client/src/images/*", dest: "_client/dist/images/" },
			   { expand: true, flatten: true, src: "_client/src/fonts/*", dest: "_client/dist/fonts/" },
		   ]
		  }
		},

		processhtml : {
			dist : {
				options : {
					process: true,
				},
				files : {
					'_client/dist/index.html': ['_client/src/index.html']
				}
			}
		},

		cachebreaker: {
			dist: {
				options: {
					match: ['giffy.min.js', 'giffy.min.css'],
					position: 'filename'
				},
				files: {
					src: ['_client/dist/index.html']
				}
			}
		},

		clean: {
			build: {
				src: [
					"_client/dist/css/giffy.css",
					"_client/dist/js/giffy.js",
					"_client/dist/js/partials.js",
				]
			},
			full : {
				src: [
					"_client/dist/*"
				]
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
	grunt.loadNpmTasks('grunt-html2js');

	grunt.registerTask(
		'build',
		'Compiles all of the assets and copies the files to the build directory.',
		[
			'copy:dist',
			'html2js:dist',
			'concat:app',
			'uglify:app',
			'less:compile',
			'cssmin',
			'processhtml:dist',
			'cachebreaker:dist',
			'clean:build'
		]
	);
}