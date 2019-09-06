# create module path, if it does not exist
mkdir -p $HOME/go/src/utropicmedia/maria_storj_interface/maria
mkdir -p $HOME/go/src/utropicmedia/maria_storj_interface/storj
# move go packages to include path
cp utropicmedia/maria_storj_interface/maria/* $HOME/go/src/utropicmedia/maria_storj_interface/maria/
cp utropicmedia/maria_storj_interface/storj/* $HOME/go/src/utropicmedia/maria_storj_interface/storj/
cp utropicmedia/maria_storj_interface/* $HOME/go/src/utropicmedia/maria_storj_interface/
