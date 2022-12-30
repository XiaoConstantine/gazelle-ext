; package
(package_declaration [
	(scoped_identifier (identifier)? @local_package)
    (identifier) @local_package
  ]
) @full_package


; import
(
 import_declaration [
 	(scoped_identifier
 		(identifier)?@class_name)
    (asterisk) @asterisk_import
  ]
)@full_import
