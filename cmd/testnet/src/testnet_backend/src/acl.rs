use candid::Principal;
use ic_cdk::caller;
use crate::{StorageRef};
use ic_cdk::storage;
use ic_stable_structures::StableBTreeMap;
use crate::PERMISSIONS;
use candid::export_service;
use ic_cdk_macros::{query, update};
use ic_stable_structures::memory_manager::MemoryId;







pub fn is_user() -> Result<(), String> {
    let caller = caller();
    let permissions = PERMISSIONS.with(|p| p.borrow().contains_key(&caller));
    
    if permissions {
        Ok(())
    } else {
        Err("Caller is not authorized.".to_string())
    }
}

pub fn is_owner() -> Result<(), String> {
    let caller = caller();
    let permissions = PERMISSIONS.with(|p| p.borrow().get(&caller));
    
    if let Some(permission) = permissions {
        if permission == true {
            Ok(())
        } else {
            Err("Caller is not the owner.".to_string())
        }
    } else {
        Err("Caller is not authorized.".to_string())
    }
}

#[query(guard = "is_owner")]
pub fn add_user(text : String) -> String {
    let res: Result<Principal, String> 
        = Principal::from_text(text).map_err(|e| e.to_string());
    if let Ok(principal) = res{
        PERMISSIONS.with(|p| p.borrow_mut().insert(principal, false));
        "success".to_string()
    } else {
        res.unwrap_err()
    }
}

#[query(guard = "is_owner")]
pub fn remove_user(principal: Principal) -> String{
    let mut permissions = PERMISSIONS.with(|p| p.borrow_mut().remove(&principal));
    "success".to_string()
}


#[query(guard = "is_user")]
pub fn get_permission(principal: Principal) -> String {
    let permissions = PERMISSIONS.with(|p| p.borrow().get(&principal));
    if let Some(permission) = permissions {
        if permission == true {
            "Owner".to_string()
        } else {
            "User".to_string()
        }
    } else {
        "Unauthorized".to_string()
    }
}

pub fn init() {
    let deployer_principal = caller();
    let mut permissions = PERMISSIONS.with(|p| p.borrow_mut().insert(deployer_principal, true));
}
